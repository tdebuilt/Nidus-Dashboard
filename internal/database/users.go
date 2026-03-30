package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/tdebuilt/nidus/internal/models"
)

// CreateUser inserts a new user into the database with the given role.
func (db *DB) CreateUser(ctx context.Context, username, passwordHash, role string) (int64, error) {
	if role == "" {
		role = "admin"
	}
	result, err := db.ExecContext(ctx,
		"INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)",
		username, passwordHash, role,
	)
	if err != nil {
		return 0, fmt.Errorf("inserting user: %w", err)
	}
	return result.LastInsertId()
}

// GetUserByUsername retrieves a user by username, or nil if not found.
func (db *DB) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	u := &models.User{}
	var totpSecret sql.NullString
	var totpEnabled int
	err := db.QueryRowContext(ctx,
		"SELECT id, username, password_hash, role, totp_secret, totp_enabled, token_version, created_at, updated_at FROM users WHERE username = ?",
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &totpSecret, &totpEnabled, &u.TokenVersion, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying user: %w", err)
	}
	if totpSecret.Valid {
		u.TOTPSecret = &totpSecret.String
	}
	u.TOTPEnabled = totpEnabled == 1
	return u, nil
}

// CountUsers returns the number of users in the database.
func (db *DB) CountUsers(ctx context.Context) (int, error) {
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

// GetUserByID retrieves a user by ID, or nil if not found.
func (db *DB) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	u := &models.User{}
	var totpSecret sql.NullString
	var totpEnabled int
	err := db.QueryRowContext(ctx,
		"SELECT id, username, password_hash, role, totp_secret, totp_enabled, token_version, created_at, updated_at FROM users WHERE id = ?",
		id,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &totpSecret, &totpEnabled, &u.TokenVersion, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying user: %w", err)
	}
	if totpSecret.Valid {
		u.TOTPSecret = &totpSecret.String
	}
	u.TOTPEnabled = totpEnabled == 1
	return u, nil
}

// SetUserTOTPSecret stores the TOTP secret for a user.
func (db *DB) SetUserTOTPSecret(ctx context.Context, userID int64, secret string) error {
	_, err := db.ExecContext(ctx,
		"UPDATE users SET totp_secret = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		secret, userID,
	)
	return err
}

// EnableUserTOTP enables TOTP for a user.
func (db *DB) EnableUserTOTP(ctx context.Context, userID int64) error {
	_, err := db.ExecContext(ctx,
		"UPDATE users SET totp_enabled = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		userID,
	)
	return err
}

// DisableUserTOTP disables TOTP and clears the secret for a user.
func (db *DB) DisableUserTOTP(ctx context.Context, userID int64) error {
	_, err := db.ExecContext(ctx,
		"UPDATE users SET totp_enabled = 0, totp_secret = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		userID,
	)
	return err
}

// ListUsers retrieves all users (without password hashes).
func (db *DB) ListUsers(ctx context.Context) ([]models.User, error) {
	rows, err := db.QueryContext(ctx,
		"SELECT id, username, role, totp_enabled, created_at, updated_at FROM users ORDER BY id",
	)
	if err != nil {
		return nil, fmt.Errorf("querying users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var totpEnabled int
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &totpEnabled, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning user: %w", err)
		}
		u.TOTPEnabled = totpEnabled == 1
		users = append(users, u)
	}
	return users, rows.Err()
}

// UpdateUserRole updates a user's role.
func (db *DB) UpdateUserRole(ctx context.Context, userID int64, role string) error {
	result, err := db.ExecContext(ctx,
		"UPDATE users SET role = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		role, userID,
	)
	if err != nil {
		return fmt.Errorf("updating user role: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// IncrementTokenVersion bumps the token version for a user, invalidating all existing JWTs.
func (db *DB) IncrementTokenVersion(ctx context.Context, userID int64) error {
	_, err := db.ExecContext(ctx, "UPDATE users SET token_version = token_version + 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?", userID)
	return err
}

// DeleteUser deletes a user by ID. Returns an error if user not found.
func (db *DB) DeleteUser(ctx context.Context, userID int64) error {
	result, err := db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// CreateInvitation creates a new invitation code.
func (db *DB) CreateInvitation(ctx context.Context, code, role string, createdBy int64, expiresAt string) error {
	_, err := db.ExecContext(ctx,
		"INSERT INTO invitations (code, role, created_by, expires_at) VALUES (?, ?, ?, ?)",
		code, role, createdBy, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("creating invitation: %w", err)
	}
	return nil
}

// GetInvitationByCode retrieves a valid (unused, not expired) invitation by code.
func (db *DB) GetInvitationByCode(ctx context.Context, code string) (*models.Invitation, error) {
	inv := &models.Invitation{}
	var usedBy sql.NullInt64
	var usedAt sql.NullTime
	err := db.QueryRowContext(ctx,
		"SELECT id, code, role, created_by, used_by, used_at, expires_at, created_at FROM invitations WHERE code = ? AND used_by IS NULL AND expires_at > CURRENT_TIMESTAMP",
		code,
	).Scan(&inv.ID, &inv.Code, &inv.Role, &inv.CreatedBy, &usedBy, &usedAt, &inv.ExpiresAt, &inv.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying invitation: %w", err)
	}
	if usedBy.Valid {
		inv.UsedBy = &usedBy.Int64
	}
	if usedAt.Valid {
		inv.UsedAt = &usedAt.Time
	}
	return inv, nil
}

// UseInvitation marks an invitation as used by a user.
func (db *DB) UseInvitation(ctx context.Context, invID, userID int64) error {
	_, err := db.ExecContext(ctx,
		"UPDATE invitations SET used_by = ?, used_at = CURRENT_TIMESTAMP WHERE id = ?",
		userID, invID,
	)
	return err
}

// ListInvitations retrieves all invitations created by a user.
func (db *DB) ListInvitations(ctx context.Context) ([]models.Invitation, error) {
	rows, err := db.QueryContext(ctx,
		"SELECT id, code, role, created_by, used_by, used_at, expires_at, created_at FROM invitations ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, fmt.Errorf("querying invitations: %w", err)
	}
	defer rows.Close()

	var invitations []models.Invitation
	for rows.Next() {
		var inv models.Invitation
		var usedBy sql.NullInt64
		var usedAt sql.NullTime
		if err := rows.Scan(&inv.ID, &inv.Code, &inv.Role, &inv.CreatedBy, &usedBy, &usedAt, &inv.ExpiresAt, &inv.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning invitation: %w", err)
		}
		if usedBy.Valid {
			inv.UsedBy = &usedBy.Int64
		}
		if usedAt.Valid {
			inv.UsedAt = &usedAt.Time
		}
		invitations = append(invitations, inv)
	}
	return invitations, rows.Err()
}

// DeleteInvitation deletes an invitation by ID.
func (db *DB) DeleteInvitation(ctx context.Context, invID int64) error {
	_, err := db.ExecContext(ctx, "DELETE FROM invitations WHERE id = ?", invID)
	return err
}

// UpdateUserUsername updates a user's username.
func (db *DB) UpdateUserUsername(ctx context.Context, userID int64, username string) error {
	result, err := db.ExecContext(ctx,
		"UPDATE users SET username = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		username, userID,
	)
	if err != nil {
		return fmt.Errorf("updating username: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// UpdateUserPassword updates a user's password hash.
func (db *DB) UpdateUserPassword(ctx context.Context, userID int64, passwordHash string) error {
	result, err := db.ExecContext(ctx,
		"UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		passwordHash, userID,
	)
	if err != nil {
		return fmt.Errorf("updating password: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// CreatePasswordReset creates a new password reset code.
func (db *DB) CreatePasswordReset(ctx context.Context, userID int64, code string, createdBy int64, expiresAt string) error {
	_, err := db.ExecContext(ctx,
		"INSERT INTO password_resets (user_id, code, created_by, expires_at) VALUES (?, ?, ?, ?)",
		userID, code, createdBy, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("creating password reset: %w", err)
	}
	return nil
}

// GetPasswordResetByCode retrieves a valid (unused, not expired) reset code.
func (db *DB) GetPasswordResetByCode(ctx context.Context, code string) (*models.PasswordReset, error) {
	reset := &models.PasswordReset{}
	var usedAt sql.NullTime
	err := db.QueryRowContext(ctx,
		"SELECT id, user_id, code, created_by, used_at, expires_at, created_at FROM password_resets WHERE code = ? AND used_at IS NULL AND expires_at > CURRENT_TIMESTAMP",
		code,
	).Scan(&reset.ID, &reset.UserID, &reset.Code, &reset.CreatedBy, &usedAt, &reset.ExpiresAt, &reset.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying password reset: %w", err)
	}
	if usedAt.Valid {
		reset.UsedAt = &usedAt.Time
	}
	return reset, nil
}

// UsePasswordReset marks a reset code as used.
func (db *DB) UsePasswordReset(ctx context.Context, resetID int64) error {
	_, err := db.ExecContext(ctx,
		"UPDATE password_resets SET used_at = CURRENT_TIMESTAMP WHERE id = ?",
		resetID,
	)
	return err
}
