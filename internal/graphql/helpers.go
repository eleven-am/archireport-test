package graphql

import (
	"context"
	"fmt"

	"github.com/eleven-am/enclave/ent"
	"github.com/eleven-am/enclave/ent/room"
	"github.com/eleven-am/enclave/ent/roommembership"
	"github.com/eleven-am/enclave/ent/user"
)

func (r *Resolver) ensureRoomAccess(ctx context.Context, roomID, userID int) error {
	_, err := r.ensureRoomMember(ctx, roomID, userID)
	return err
}

func (r *Resolver) ensureRoomMember(ctx context.Context, roomID, userID int) (*ent.RoomMembership, error) {
	membership, err := r.Client.RoomMembership.Query().
		Where(
			roommembership.HasRoomWith(room.ID(roomID)),
			roommembership.HasUserWith(user.IDEQ(userID)),
		).
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, ErrForbidden
	}
	return membership, err
}

func (r *Resolver) ensureRoomAdmin(ctx context.Context, roomID, userID int) error {
	membership, err := r.ensureRoomMember(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if membership.Role == roommembership.RoleOwner || membership.Role == roommembership.RoleAdmin {
		return nil
	}
	return ErrForbidden
}

func (r *Resolver) ensureRoomOwner(ctx context.Context, roomID, userID int) error {
	membership, err := r.ensureRoomMember(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if membership.Role == roommembership.RoleOwner {
		return nil
	}
	return ErrForbidden
}

func rollbackOnError(tx *ent.Tx, errPtr *error) {
	if errPtr == nil || *errPtr == nil {
		return
	}
	if rollbackErr := tx.Rollback(); rollbackErr != nil {
		*errPtr = fmt.Errorf("rollback failed: %w (original error: %v)", rollbackErr, *errPtr)
	}
}
