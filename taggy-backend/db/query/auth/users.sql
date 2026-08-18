-- =========================================
-- GET
-- =========================================

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1
    AND is_deleted = FALSE;


-- name: GetUserByPublicID :one
SELECT *
FROM users
WHERE public_id = $1
    AND is_deleted = FALSE;


-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1
    AND is_deleted = FALSE;


-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE username = $1
    AND is_deleted = FALSE;


-- =========================================
-- CREATE
-- =========================================

-- name: CreateUser :one
INSERT INTO users (
    email,
    username,
    name,
    profile_picture_url,
    bio,
    subscription,
    email_verified
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
)
RETURNING *;


-- =========================================
-- UPDATE
-- =========================================

-- name: UpdateUserProfile :one
UPDATE users
SET
    name = $2,
    bio = $3,
    profile_picture_url = $4,
    updated_at = NOW()
WHERE id = $1
    AND is_deleted = FALSE
RETURNING *;


-- name: VerifyUserEmail :one
UPDATE users
SET
    email_verified = TRUE,
    updated_at = NOW()
WHERE id = $1
    AND is_deleted = FALSE
RETURNING *;


-- name: UpdateSubscription :one
UPDATE users
SET
    subscription = $2,
    updated_at = NOW()
WHERE id = $1
    AND is_deleted = FALSE
RETURNING *;


-- =========================================
-- DELETE (Soft Delete)
-- =========================================

-- name: SoftDeleteUser :exec
UPDATE users
SET
    is_deleted = TRUE,
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1;


-- name: RestoreUser :exec
UPDATE users
SET
    is_deleted = FALSE,
    deleted_at = NULL,
    updated_at = NOW()
WHERE id = $1;


-- =========================================
-- EXISTS
-- =========================================

-- name: EmailExists :one
SELECT EXISTS (
    SELECT 1
    FROM users
    WHERE email = $1
        AND is_deleted = FALSE
);


-- name: UsernameExists :one
SELECT EXISTS (
    SELECT 1
    FROM users
    WHERE username = $1
        AND is_deleted = FALSE
);
