package audio

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

const (
	emptyRoomTTL      = 3 * time.Minute
	emptyRoomInterval = 30 * time.Second
)

// RunEmptyRoomSweeper deletes ACTIVE rooms with no participants for longer than emptyRoomTTL.
// Blocks until ctx is cancelled.
func (s *Service) RunEmptyRoomSweeper(ctx context.Context) {
	ticker := time.NewTicker(emptyRoomInterval)
	defer ticker.Stop()

	s.log.Info().
		Dur("ttl", emptyRoomTTL).
		Dur("interval", emptyRoomInterval).
		Msg("audio empty-room sweeper started")

	s.sweepEmptyRooms(ctx)

	for {
		select {
		case <-ctx.Done():
			s.log.Info().Msg("audio empty-room sweeper stopped")
			return
		case <-ticker.C:
			s.sweepEmptyRooms(ctx)
		}
	}
}

func (s *Service) sweepEmptyRooms(ctx context.Context) {
	cutoff := time.Now().UTC().Add(-emptyRoomTTL)
	rooms, err := s.repo.ListStaleEmptyActiveAudioRooms(ctx, pgtype.Timestamptz{
		Time:  cutoff,
		Valid: true,
	})
	if err != nil {
		s.log.Error().Err(err).Msg("list stale empty audio rooms failed")
		return
	}

	for _, room := range rooms {
		s.deleteLiveKitRoom(ctx, room.LivekitRoomName)
		if err := s.repo.DeleteAudioRoom(ctx, room.ID); err != nil {
			s.log.Error().
				Err(err).
				Str("room_id", room.PublicID.String()).
				Msg("delete empty audio room failed")
			continue
		}
		s.log.Info().
			Str("room_id", room.PublicID.String()).
			Str("livekit_room_name", room.LivekitRoomName).
			Msg("deleted audio room empty for over 3 minutes")
	}
}
