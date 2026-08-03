package password

// Config controls how passwords are hashed.
//
// Cost is the bcrypt "work factor". Higher cost = more CPU time per hash,
// which makes brute-force attacks slower for an attacker but also makes
// login/register slightly slower for legitimate users.
//
// Typical values:
//   - 10–12 for most web apps (default in many libraries is 10)
//   - increase over time as hardware gets faster
//
// Example: cost 12 means hashing takes ~250ms on modern hardware.
// An attacker trying billions of guesses pays that cost every time.
type Config struct {
	Cost int
}
