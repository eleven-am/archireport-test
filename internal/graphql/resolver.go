package graphql

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/mitchellh/mapstructure"

	"github.com/eleven-am/enclave/ent"
	"github.com/eleven-am/enclave/ent/calllog"
	"github.com/eleven-am/enclave/ent/callparticipant"
	"github.com/eleven-am/enclave/ent/contact"
	"github.com/eleven-am/enclave/ent/favourite"
	"github.com/eleven-am/enclave/ent/media"
	"github.com/eleven-am/enclave/ent/message"
	"github.com/eleven-am/enclave/ent/notification"
	"github.com/eleven-am/enclave/ent/room"
	"github.com/eleven-am/enclave/ent/roommembership"
	"github.com/eleven-am/enclave/ent/user"
	"github.com/eleven-am/enclave/internal/auth"
)

// Resolver encapsulates access to ent.Client for GraphQL handlers.
type NotificationListener func(context.Context, int, *ent.Notification)

type Resolver struct {
	Client                *ent.Client
	userObj               *graphql.Object
	roomObj               *graphql.Object
	roomMembershipObj     *graphql.Object
	messageObj            *graphql.Object
	mediaObj              *graphql.Object
	contactObj            *graphql.Object
	favouriteObj          *graphql.Object
	callLogObj            *graphql.Object
	callParticipantObj    *graphql.Object
	notificationObj       *graphql.Object
	notificationBroker    *notificationBroker
	notificationListeners []NotificationListener
}

// ErrUnauthorized indicates the caller is not authorized to perform an action.
var ErrUnauthorized = errors.New("unauthorized")

// ErrForbidden indicates the caller is authenticated but not permitted.
var ErrForbidden = errors.New("forbidden")

// NewSchema constructs the GraphQL schema with resolvers backed by ent.
func NewSchema(client *ent.Client) (graphql.Schema, *Resolver, error) {
	r := &Resolver{Client: client, notificationBroker: newNotificationBroker()}
	schemaConfig := graphql.SchemaConfig{
		Query:        graphql.NewObject(r.queryFields()),
		Mutation:     graphql.NewObject(r.mutationFields()),
		Subscription: graphql.NewObject(r.subscriptionFields()),
	}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		return graphql.Schema{}, nil, err
	}
	return schema, r, nil
}

func (r *Resolver) RegisterNotificationListener(fn NotificationListener) {
	if fn == nil {
		return
	}
	r.notificationListeners = append(r.notificationListeners, fn)
}

func (r *Resolver) publishNotification(ctx context.Context, userID int, n *ent.Notification) {
	r.notificationBroker.Publish(ctx, n)
	for _, listener := range r.notificationListeners {
		listener(ctx, userID, n)
	}
}

func (r *Resolver) queryFields() graphql.ObjectConfig {
	return graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"users": &graphql.Field{
				Type: graphql.NewList(r.userType()),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					users, err := r.Client.User.Query().Order(ent.Asc(user.FieldUsername)).All(p.Context)
					return users, err
				},
			},
			"user": &graphql.Field{
				Type: r.userType(),
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, err := decodeID(p.Args["id"])
					if err != nil {
						return nil, err
					}
					return r.Client.User.Get(p.Context, id)
				},
			},
			"rooms": &graphql.Field{
				Type: graphql.NewList(r.roomType()),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					memberships, err := r.Client.RoomMembership.Query().
						Where(roommembership.HasUserWith(user.ID(uid))).
						WithRoom(func(q *ent.RoomQuery) {
							q.WithOwner()
						}).
						All(p.Context)
					if err != nil {
						return nil, err
					}
					rooms := make([]*ent.Room, 0, len(memberships))
					for _, m := range memberships {
						if room := m.Edges.Room; room != nil {
							rooms = append(rooms, room)
						}
					}
					sort.SliceStable(rooms, func(i, j int) bool {
						return rooms[i].CreatedAt.Before(rooms[j].CreatedAt)
					})
					return rooms, nil
				},
			},
			"room": &graphql.Field{
				Type: r.roomType(),
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					id, err := decodeID(p.Args["id"])
					if err != nil {
						return nil, err
					}
					if err := r.ensureRoomAccess(p.Context, id, uid); err != nil {
						return nil, err
					}
					return r.Client.Room.Query().Where(room.ID(id)).WithOwner().Only(p.Context)
				},
			},
			"messages": &graphql.Field{
				Type: graphql.NewList(r.messageType()),
				Args: graphql.FieldConfigArgument{
					"roomId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					roomID, err := decodeID(p.Args["roomId"])
					if err != nil {
						return nil, err
					}
					if err := r.ensureRoomAccess(p.Context, roomID, uid); err != nil {
						return nil, err
					}
					return r.Client.Message.Query().
						Where(message.HasRoomWith(room.ID(roomID))).
						WithSender().
						Order(ent.Asc(message.FieldCreatedAt)).
						All(p.Context)
				},
			},
			"notifications": &graphql.Field{
				Type: graphql.NewList(r.notificationType()),
				Args: graphql.FieldConfigArgument{
					"unreadOnly": &graphql.ArgumentConfig{Type: graphql.Boolean},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					query := r.Client.Notification.Query().
						Where(notification.HasRecipientWith(user.ID(uid))).
						Order(ent.Desc(notification.FieldCreatedAt)).
						WithRecipient().
						WithRoom().
						WithMessage()
					if unreadOnly, ok := p.Args["unreadOnly"].(bool); ok && unreadOnly {
						query = query.Where(notification.ReadEQ(false))
					}
					return query.All(p.Context)
				},
			},
			"notification": &graphql.Field{
				Type: r.notificationType(),
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					id, err := decodeID(p.Args["id"])
					if err != nil {
						return nil, err
					}
					return r.Client.Notification.Query().
						Where(notification.ID(id), notification.HasRecipientWith(user.ID(uid))).
						WithRecipient().
						WithRoom().
						WithMessage().
						Only(p.Context)
				},
			},
			"contacts": &graphql.Field{
				Type: graphql.NewList(r.contactType()),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					return r.Client.Contact.Query().
						Where(contact.HasOwnerWith(user.ID(uid))).
						WithOwner().
						WithContact().
						Order(ent.Asc(contact.FieldCreatedAt)).
						All(p.Context)
				},
			},
			"favourites": &graphql.Field{
				Type: graphql.NewList(r.favouriteType()),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					return r.Client.Favourite.Query().
						Where(favourite.HasUserWith(user.ID(uid))).
						WithUser().
						WithRoom().
						All(p.Context)
				},
			},
			"callLogs": &graphql.Field{
				Type: graphql.NewList(r.callLogType()),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					return r.Client.CallLog.Query().
						Where(calllog.Or(
							calllog.HasInitiatorWith(user.ID(uid)),
							calllog.HasParticipantsWith(callparticipant.HasParticipantWith(user.ID(uid))),
						)).
						WithInitiator().
						WithRoom().
						WithParticipants(func(q *ent.CallParticipantQuery) {
							q.WithParticipant()
						}).
						Order(ent.Desc(calllog.FieldStartedAt)).
						All(p.Context)
				},
			},
		},
	}
}

func (r *Resolver) mutationFields() graphql.ObjectConfig {
	return graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createUser": &graphql.Field{
				Type: r.userType(),
				Args: graphql.FieldConfigArgument{
					"username":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"displayName": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"email":       &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"avatarUrl":   &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					builder := r.Client.User.Create().
						SetUsername(p.Args["username"].(string)).
						SetDisplayName(p.Args["displayName"].(string)).
						SetEmail(p.Args["email"].(string))
					if avatar, ok := p.Args["avatarUrl"].(string); ok && avatar != "" {
						builder.SetAvatarURL(avatar)
					}
					return builder.Save(p.Context)
				},
			},
			"updateUser": &graphql.Field{
				Type: r.userType(),
				Args: graphql.FieldConfigArgument{
					"id":          &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"displayName": &graphql.ArgumentConfig{Type: graphql.String},
					"avatarUrl":   &graphql.ArgumentConfig{Type: graphql.String},
					"lastSeenAt":  &graphql.ArgumentConfig{Type: graphql.DateTime},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					id, err := decodeID(p.Args["id"])
					if err != nil {
						return nil, err
					}
					if uid != id {
						return nil, ErrForbidden
					}
					builder := r.Client.User.UpdateOneID(id)
					if v, ok := p.Args["displayName"].(string); ok {
						builder.SetDisplayName(v)
					}
					if v, ok := p.Args["avatarUrl"].(string); ok {
						if v == "" {
							builder.ClearAvatarURL()
						} else {
							builder.SetAvatarURL(v)
						}
					}
					if v, ok := p.Args["lastSeenAt"].(time.Time); ok {
						builder.SetLastSeenAt(v)
					}
					builder.SetUpdatedAt(time.Now())
					return builder.Save(p.Context)
				},
			},
			"deleteUser": &graphql.Field{
				Type: graphql.Boolean,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					id, err := decodeID(p.Args["id"])
					if err != nil {
						return nil, err
					}
					if uid != id {
						return nil, ErrForbidden
					}
					return true, r.Client.User.DeleteOneID(id).Exec(p.Context)
				},
			},
			"createRoom": &graphql.Field{
				Type: r.roomType(),
				Args: graphql.FieldConfigArgument{
					"name":           &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"description":    &graphql.ArgumentConfig{Type: graphql.String},
					"isPrivate":      &graphql.ArgumentConfig{Type: graphql.Boolean},
					"participantIds": &graphql.ArgumentConfig{Type: graphql.NewList(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					description := ""
					if v, ok := p.Args["description"].(string); ok {
						description = v
					}
					isPrivate := false
					if v, ok := p.Args["isPrivate"].(bool); ok {
						isPrivate = v
					}
					tx, err := r.Client.Tx(p.Context)
					if err != nil {
						return nil, err
					}
					defer rollbackOnError(tx, &err)

					roomBuilder := tx.Room.Create().
						SetName(p.Args["name"].(string)).
						SetDescription(description).
						SetIsPrivate(isPrivate).
						SetOwnerID(uid)

					participantIDs := decodeIDList(p.Args["participantIds"])
					nonOwnerCount := 0
					for _, pid := range participantIDs {
						if pid != uid {
							nonOwnerCount++
						}
					}
					direct := nonOwnerCount == 1
					roomBuilder.SetIsDirect(direct)

					newRoom, err := roomBuilder.Save(p.Context)
					if err != nil {
						return nil, err
					}

					_, err = tx.RoomMembership.Create().
						SetRoom(newRoom).
						SetUserID(uid).
						SetRole(roommembership.RoleOwner).
						Save(p.Context)
					if err != nil {
						return nil, err
					}

					for _, pid := range participantIDs {
						role := roommembership.RoleMember
						if direct && pid != uid {
							role = roommembership.RoleAdmin
						}
						_, err = tx.RoomMembership.Create().
							SetRoom(newRoom).
							SetUserID(pid).
							SetRole(role).
							Save(p.Context)
						if err != nil {
							return nil, err
						}
					}

					if err = tx.Commit(); err != nil {
						return nil, err
					}
					return r.Client.Room.Query().Where(room.ID(newRoom.ID)).WithOwner().Only(p.Context)
				},
			},
			"updateRoom": &graphql.Field{
				Type: r.roomType(),
				Args: graphql.FieldConfigArgument{
					"id":          &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"name":        &graphql.ArgumentConfig{Type: graphql.String},
					"description": &graphql.ArgumentConfig{Type: graphql.String},
					"isPrivate":   &graphql.ArgumentConfig{Type: graphql.Boolean},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					roomID, err := decodeID(p.Args["id"])
					if err != nil {
						return nil, err
					}
					if err := r.ensureRoomAdmin(p.Context, roomID, uid); err != nil {
						return nil, err
					}
					builder := r.Client.Room.UpdateOneID(roomID)
					if v, ok := p.Args["name"].(string); ok {
						builder.SetName(v)
					}
					if v, ok := p.Args["description"].(string); ok {
						builder.SetDescription(v)
					}
					if v, ok := p.Args["isPrivate"].(bool); ok {
						builder.SetIsPrivate(v)
					}
					builder.SetUpdatedAt(time.Now())
					if err := builder.Exec(p.Context); err != nil {
						return nil, err
					}
					return r.Client.Room.Query().Where(room.ID(roomID)).WithOwner().Only(p.Context)
				},
			},
			"deleteRoom": &graphql.Field{
				Type: graphql.Boolean,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					roomID, err := decodeID(p.Args["id"])
					if err != nil {
						return nil, err
					}
					if err := r.ensureRoomOwner(p.Context, roomID, uid); err != nil {
						return nil, err
					}
					return true, r.Client.Room.DeleteOneID(roomID).Exec(p.Context)
				},
			},
			"addRoomMembers": &graphql.Field{
				Type: graphql.NewList(r.roomMembershipType()),
				Args: graphql.FieldConfigArgument{
					"roomId":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"memberIds": &graphql.ArgumentConfig{Type: graphql.NewList(graphql.ID)},
					"role":      &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					roomID, err := decodeID(p.Args["roomId"])
					if err != nil {
						return nil, err
					}
					if err := r.ensureRoomAdmin(p.Context, roomID, uid); err != nil {
						return nil, err
					}
					role := roommembership.RoleMember
					if v, ok := p.Args["role"].(string); ok && v != "" {
						role = roommembership.Role(v)
					}
					tx, err := r.Client.Tx(p.Context)
					if err != nil {
						return nil, err
					}
					defer rollbackOnError(tx, &err)

					members := []*ent.RoomMembership{}
					for _, mid := range decodeIDList(p.Args["memberIds"]) {
						m, err := tx.RoomMembership.Create().
							SetRoomID(roomID).
							SetUserID(mid).
							SetRole(role).
							Save(p.Context)
						if err != nil {
							return nil, err
						}
						members = append(members, m)
					}
					if err = tx.Commit(); err != nil {
						return nil, err
					}
					return r.Client.RoomMembership.Query().
						Where(roommembership.HasRoomWith(room.ID(roomID)), roommembership.HasUserWith(user.IDIn(decodeIDList(p.Args["memberIds"])...))).
						WithRoom().
						WithUser().
						All(p.Context)
				},
			},
			"updateRoomMembership": &graphql.Field{
				Type: r.roomMembershipType(),
				Args: graphql.FieldConfigArgument{
					"roomId":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"memberId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"role":     &graphql.ArgumentConfig{Type: graphql.String},
					"canPost":  &graphql.ArgumentConfig{Type: graphql.Boolean},
					"canCall":  &graphql.ArgumentConfig{Type: graphql.Boolean},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					roomID, err := decodeID(p.Args["roomId"])
					if err != nil {
						return nil, err
					}
					memberID, err := decodeID(p.Args["memberId"])
					if err != nil {
						return nil, err
					}
					if err := r.ensureRoomAdmin(p.Context, roomID, uid); err != nil {
						return nil, err
					}
					builder := r.Client.RoomMembership.Update().
						Where(roommembership.HasRoomWith(room.ID(roomID)), roommembership.HasUserWith(user.ID(memberID)))
					if v, ok := p.Args["role"].(string); ok && v != "" {
						builder.SetRole(roommembership.Role(v))
					}
					if v, ok := p.Args["canPost"].(bool); ok {
						builder.SetCanPost(v)
					}
					if v, ok := p.Args["canCall"].(bool); ok {
						builder.SetCanCall(v)
					}
					builder.SetUpdatedAt(time.Now())
					if _, err := builder.Save(p.Context); err != nil {
						return nil, err
					}
					return r.Client.RoomMembership.Query().
						Where(roommembership.HasRoomWith(room.ID(roomID)), roommembership.HasUserWith(user.ID(memberID))).
						WithRoom().
						WithUser().
						Only(p.Context)
				},
			},
			"removeRoomMember": &graphql.Field{
				Type: graphql.Boolean,
				Args: graphql.FieldConfigArgument{
					"roomId":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"memberId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					roomID, err := decodeID(p.Args["roomId"])
					if err != nil {
						return nil, err
					}
					memberID, err := decodeID(p.Args["memberId"])
					if err != nil {
						return nil, err
					}
					if err := r.ensureRoomAdmin(p.Context, roomID, uid); err != nil {
						return nil, err
					}
					_, err = r.Client.RoomMembership.Delete().
						Where(roommembership.HasRoomWith(room.ID(roomID)), roommembership.HasUserWith(user.ID(memberID))).
						Exec(p.Context)
					if err != nil {
						return nil, err
					}
					return true, nil
				},
			},
			"createMessage": &graphql.Field{
				Type: r.messageType(),
				Args: graphql.FieldConfigArgument{
					"roomId":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"cipherText":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"contentType": &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					roomID, err := decodeID(p.Args["roomId"])
					if err != nil {
						return nil, err
					}
					membership, err := r.ensureRoomMember(p.Context, roomID, uid)
					if err != nil {
						return nil, err
					}
					if !membership.CanPost {
						return nil, ErrForbidden
					}
					builder := r.Client.Message.Create().
						SetRoomID(roomID).
						SetSenderID(uid).
						SetCipherText(p.Args["cipherText"].(string))
					if v, ok := p.Args["contentType"].(string); ok && v != "" {
						builder.SetContentType(v)
					}
					return builder.Save(p.Context)
				},
			},
			"updateMessage": &graphql.Field{
				Type: r.messageType(),
				Args: graphql.FieldConfigArgument{
					"id":          &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"cipherText":  &graphql.ArgumentConfig{Type: graphql.String},
					"contentType": &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					id, err := decodeID(p.Args["id"])
					if err != nil {
						return nil, err
					}
					msg, err := r.Client.Message.Query().Where(message.IDEQ(id)).
						WithRoom().
						WithSender().
						Only(p.Context)
					if err != nil {
						return nil, err
					}
					roomEdge := msg.Edges.Room
					if roomEdge == nil {
						return nil, fmt.Errorf("message missing room relationship")
					}
					membership, err := r.ensureRoomMember(p.Context, roomEdge.ID, uid)
					if err != nil {
						return nil, err
					}
					senderEdge := msg.Edges.Sender
					if senderEdge == nil {
						return nil, fmt.Errorf("message missing sender relationship")
					}
					if senderEdge.ID != uid && membership.Role == roommembership.RoleMember {
						return nil, ErrForbidden
					}
					builder := r.Client.Message.UpdateOneID(id).SetEdited(true)
					if v, ok := p.Args["cipherText"].(string); ok {
						builder.SetCipherText(v)
					}
					if v, ok := p.Args["contentType"].(string); ok {
						builder.SetContentType(v)
					}
					builder.SetUpdatedAt(time.Now())
					if err := builder.Exec(p.Context); err != nil {
						return nil, err
					}
					return r.Client.Message.Query().
						Where(message.ID(id)).
						WithSender().
						Only(p.Context)
				},
			},
			"deleteMessage": &graphql.Field{
				Type: graphql.Boolean,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					id, err := decodeID(p.Args["id"])
					if err != nil {
						return nil, err
					}
					msg, err := r.Client.Message.Query().Where(message.IDEQ(id)).
						WithSender().
						WithRoom().
						Only(p.Context)
					if err != nil {
						return nil, err
					}
					roomEdge := msg.Edges.Room
					if roomEdge == nil {
						return nil, fmt.Errorf("message missing room relationship")
					}
					membership, err := r.ensureRoomMember(p.Context, roomEdge.ID, uid)
					if err != nil {
						return nil, err
					}
					senderEdge := msg.Edges.Sender
					if senderEdge == nil {
						return nil, fmt.Errorf("message missing sender relationship")
					}
					if senderEdge.ID != uid && membership.Role == roommembership.RoleMember {
						return nil, ErrForbidden
					}
					return true, r.Client.Message.DeleteOneID(id).Exec(p.Context)
				},
			},
			"createNotification": &graphql.Field{
				Type: r.notificationType(),
				Args: graphql.FieldConfigArgument{
					"recipientId":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"kind":             &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"cipherText":       &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"encryptionScheme": &graphql.ArgumentConfig{Type: graphql.String},
					"roomId":           &graphql.ArgumentConfig{Type: graphql.ID},
					"messageId":        &graphql.ArgumentConfig{Type: graphql.ID},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					recipientID, err := decodeID(p.Args["recipientId"])
					if err != nil {
						return nil, err
					}
					var roomID *int
					if roomArg, ok := p.Args["roomId"]; ok && roomArg != nil {
						rid, err := decodeID(roomArg)
						if err != nil {
							return nil, err
						}
						roomID = &rid
					}
					var messageID *int
					if msgArg, ok := p.Args["messageId"]; ok && msgArg != nil {
						mid, err := decodeID(msgArg)
						if err != nil {
							return nil, err
						}
						message, err := r.Client.Message.Query().Where(message.IDEQ(mid)).WithRoom().Only(p.Context)
						if err != nil {
							return nil, err
						}
						roomEdge := message.Edges.Room
						if roomEdge == nil {
							return nil, fmt.Errorf("message missing room relationship")
						}
						if roomID != nil && *roomID != roomEdge.ID {
							return nil, fmt.Errorf("message does not belong to provided room")
						}
						roomID = &roomEdge.ID
						messageID = &mid
					}
					if roomID != nil {
						if _, err := r.ensureRoomMember(p.Context, *roomID, recipientID); err != nil {
							return nil, err
						}
						if recipientID != uid {
							if err := r.ensureRoomAdmin(p.Context, *roomID, uid); err != nil {
								return nil, err
							}
						} else {
							if _, err := r.ensureRoomMember(p.Context, *roomID, uid); err != nil {
								return nil, err
							}
						}
					} else if recipientID != uid {
						return nil, ErrForbidden
					}
					builder := r.Client.Notification.Create().
						SetRecipientID(recipientID).
						SetKind(p.Args["kind"].(string)).
						SetCipherText(p.Args["cipherText"].(string))
					if scheme, ok := p.Args["encryptionScheme"].(string); ok && scheme != "" {
						builder.SetEncryptionScheme(scheme)
					}
					if roomID != nil {
						builder.SetRoomID(*roomID)
					}
					if messageID != nil {
						builder.SetMessageID(*messageID)
					}
					savedNotification, err := builder.Save(p.Context)
					if err != nil {
						return nil, err
					}
					enriched, err := r.Client.Notification.Query().
						Where(notification.IDEQ(savedNotification.ID)).
						WithRecipient().
						WithRoom().
						WithMessage().
						Only(p.Context)
					if err != nil {
						return nil, err
					}
					r.publishNotification(p.Context, recipientID, enriched)
					return enriched, nil
				},
			},
			"updateNotification": &graphql.Field{
				Type: r.notificationType(),
				Args: graphql.FieldConfigArgument{
					"id":               &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"kind":             &graphql.ArgumentConfig{Type: graphql.String},
					"cipherText":       &graphql.ArgumentConfig{Type: graphql.String},
					"encryptionScheme": &graphql.ArgumentConfig{Type: graphql.String},
					"read":             &graphql.ArgumentConfig{Type: graphql.Boolean},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					id, err := decodeID(p.Args["id"])
					if err != nil {
						return nil, err
					}
					if _, err := r.Client.Notification.Query().
						Where(notification.ID(id), notification.HasRecipientWith(user.ID(uid))).
						Only(p.Context); err != nil {
						return nil, err
					}
					builder := r.Client.Notification.UpdateOneID(id)
					if v, ok := p.Args["kind"].(string); ok {
						builder.SetKind(v)
					}
					if v, ok := p.Args["cipherText"].(string); ok {
						builder.SetCipherText(v)
					}
					if v, ok := p.Args["encryptionScheme"].(string); ok {
						builder.SetEncryptionScheme(v)
					}
					if v, ok := p.Args["read"].(bool); ok {
						builder.SetRead(v)
					}
					builder.SetUpdatedAt(time.Now())
					if err := builder.Exec(p.Context); err != nil {
						return nil, err
					}
					updated, err := r.Client.Notification.Query().
						Where(notification.ID(id)).
						WithRecipient().
						WithRoom().
						WithMessage().
						Only(p.Context)
					if err != nil {
						return nil, err
					}
					r.publishNotification(p.Context, uid, updated)
					return updated, nil
				},
			},
			"deleteNotification": &graphql.Field{
				Type: graphql.Boolean,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					id, err := decodeID(p.Args["id"])
					if err != nil {
						return nil, err
					}
					if _, err := r.Client.Notification.Query().
						Where(notification.ID(id), notification.HasRecipientWith(user.ID(uid))).
						Only(p.Context); err != nil {
						return nil, err
					}
					return true, r.Client.Notification.DeleteOneID(id).Exec(p.Context)
				},
			},
			"createMedia": &graphql.Field{
				Type: r.mediaType(),
				Args: graphql.FieldConfigArgument{
					"messageId":   &graphql.ArgumentConfig{Type: graphql.ID},
					"filename":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"contentType": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"storagePath": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"checksum":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"sizeBytes":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					builder := r.Client.Media.Create().
						SetFilename(p.Args["filename"].(string)).
						SetContentType(p.Args["contentType"].(string)).
						SetStoragePath(p.Args["storagePath"].(string)).
						SetChecksum(p.Args["checksum"].(string)).
						SetSizeBytes(int64(p.Args["sizeBytes"].(int))).
						SetUploaderID(uid)
					if msgArg, ok := p.Args["messageId"]; ok && msgArg != nil {
						msgID, err := decodeID(msgArg)
						if err != nil {
							return nil, err
						}
						msg, err := r.Client.Message.Query().Where(message.IDEQ(msgID)).
							WithSender().
							WithRoom().
							Only(p.Context)
						if err != nil {
							return nil, err
						}
						senderEdge := msg.Edges.Sender
						if senderEdge == nil {
							return nil, fmt.Errorf("message missing sender relationship")
						}
						if senderEdge.ID != uid {
							roomEdge := msg.Edges.Room
							if roomEdge == nil {
								return nil, fmt.Errorf("message missing room relationship")
							}
							membership, err := r.ensureRoomMember(p.Context, roomEdge.ID, uid)
							if err != nil {
								return nil, err
							}
							if membership.Role == roommembership.RoleMember {
								return nil, ErrForbidden
							}
						}
						builder.SetMessageID(msgID)
					}
					return builder.Save(p.Context)
				},
			},
			"deleteMedia": &graphql.Field{
				Type: graphql.Boolean,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					id, err := decodeID(p.Args["id"])
					if err != nil {
						return nil, err
					}
					mediaItem, err := r.Client.Media.Query().Where(media.IDEQ(id)).WithUploader().Only(p.Context)
					if err != nil {
						return nil, err
					}
					uploader := mediaItem.Edges.Uploader
					if uploader == nil {
						return nil, fmt.Errorf("media missing uploader relationship")
					}
					if uploader.ID != uid {
						return nil, ErrForbidden
					}
					return true, r.Client.Media.DeleteOneID(id).Exec(p.Context)
				},
			},
			"createContact": &graphql.Field{
				Type: r.contactType(),
				Args: graphql.FieldConfigArgument{
					"contactId":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"alias":       &graphql.ArgumentConfig{Type: graphql.String},
					"isFavourite": &graphql.ArgumentConfig{Type: graphql.Boolean},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					contactID, err := decodeID(p.Args["contactId"])
					if err != nil {
						return nil, err
					}
					alias := ""
					if v, ok := p.Args["alias"].(string); ok {
						alias = v
					}
					isFav := false
					if v, ok := p.Args["isFavourite"].(bool); ok {
						isFav = v
					}
					return r.Client.Contact.Create().
						SetOwnerID(uid).
						SetContactID(contactID).
						SetAlias(alias).
						SetIsFavourite(isFav).
						Save(p.Context)
				},
			},
			"updateContact": &graphql.Field{
				Type: r.contactType(),
				Args: graphql.FieldConfigArgument{
					"id":          &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"alias":       &graphql.ArgumentConfig{Type: graphql.String},
					"isFavourite": &graphql.ArgumentConfig{Type: graphql.Boolean},
					"isBlocked":   &graphql.ArgumentConfig{Type: graphql.Boolean},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					id, err := decodeID(p.Args["id"])
					if err != nil {
						return nil, err
					}
					contactItem, err := r.Client.Contact.Query().Where(contact.IDEQ(id)).WithOwner().Only(p.Context)
					if err != nil {
						return nil, err
					}
					owner := contactItem.Edges.Owner
					if owner == nil {
						return nil, fmt.Errorf("contact missing owner relationship")
					}
					if owner.ID != uid {
						return nil, ErrForbidden
					}
					builder := r.Client.Contact.UpdateOneID(id)
					if v, ok := p.Args["alias"].(string); ok {
						builder.SetAlias(v)
					}
					if v, ok := p.Args["isFavourite"].(bool); ok {
						builder.SetIsFavourite(v)
					}
					if v, ok := p.Args["isBlocked"].(bool); ok {
						builder.SetIsBlocked(v)
					}
					builder.SetUpdatedAt(time.Now())
					if err := builder.Exec(p.Context); err != nil {
						return nil, err
					}
					return r.Client.Contact.Query().Where(contact.ID(id)).WithOwner().WithContact().Only(p.Context)
				},
			},
			"deleteContact": &graphql.Field{
				Type: graphql.Boolean,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					id, err := decodeID(p.Args["id"])
					if err != nil {
						return nil, err
					}
					contactItem, err := r.Client.Contact.Query().Where(contact.IDEQ(id)).WithOwner().Only(p.Context)
					if err != nil {
						return nil, err
					}
					owner := contactItem.Edges.Owner
					if owner == nil {
						return nil, fmt.Errorf("contact missing owner relationship")
					}
					if owner.ID != uid {
						return nil, ErrForbidden
					}
					return true, r.Client.Contact.DeleteOneID(id).Exec(p.Context)
				},
			},
			"createFavourite": &graphql.Field{
				Type: r.favouriteType(),
				Args: graphql.FieldConfigArgument{
					"roomId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					roomID, err := decodeID(p.Args["roomId"])
					if err != nil {
						return nil, err
					}
					if err := r.ensureRoomAccess(p.Context, roomID, uid); err != nil {
						return nil, err
					}
					return r.Client.Favourite.Create().SetRoomID(roomID).SetUserID(uid).Save(p.Context)
				},
			},
			"deleteFavourite": &graphql.Field{
				Type: graphql.Boolean,
				Args: graphql.FieldConfigArgument{
					"roomId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					roomID, err := decodeID(p.Args["roomId"])
					if err != nil {
						return nil, err
					}
					_, err = r.Client.Favourite.Delete().
						Where(favourite.HasRoomWith(room.ID(roomID)), favourite.HasUserWith(user.ID(uid))).
						Exec(p.Context)
					if err != nil {
						return nil, err
					}
					return true, nil
				},
			},
			"createCallLog": &graphql.Field{
				Type: r.callLogType(),
				Args: graphql.FieldConfigArgument{
					"roomId":         &graphql.ArgumentConfig{Type: graphql.ID},
					"status":         &graphql.ArgumentConfig{Type: graphql.String},
					"participantIds": &graphql.ArgumentConfig{Type: graphql.NewList(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					builder := r.Client.CallLog.Create().SetInitiatorID(uid)
					if roomArg, ok := p.Args["roomId"]; ok && roomArg != nil {
						roomID, err := decodeID(roomArg)
						if err != nil {
							return nil, err
						}
						if err := r.ensureRoomAccess(p.Context, roomID, uid); err != nil {
							return nil, err
						}
						builder.SetRoomID(roomID)
					}
					if status, ok := p.Args["status"].(string); ok && status != "" {
						builder.SetStatus(calllog.Status(status))
					}
					callEntry, err := builder.Save(p.Context)
					if err != nil {
						return nil, err
					}
					for _, pid := range decodeIDList(p.Args["participantIds"]) {
						_, err := r.Client.CallParticipant.Create().
							SetCall(callEntry).
							SetParticipantID(pid).
							Save(p.Context)
						if err != nil {
							return nil, err
						}
					}
					return r.Client.CallLog.Query().
						Where(calllog.IDEQ(callEntry.ID)).
						WithInitiator().
						WithParticipants(func(q *ent.CallParticipantQuery) {
							q.WithParticipant()
						}).
						WithRoom().
						Only(p.Context)
				},
			},
			"updateCallLog": &graphql.Field{
				Type: r.callLogType(),
				Args: graphql.FieldConfigArgument{
					"id":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"status":  &graphql.ArgumentConfig{Type: graphql.String},
					"endedAt": &graphql.ArgumentConfig{Type: graphql.DateTime},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					id, err := decodeID(p.Args["id"])
					if err != nil {
						return nil, err
					}
					callEntry, err := r.Client.CallLog.Query().Where(calllog.IDEQ(id)).WithInitiator().WithRoom().Only(p.Context)
					if err != nil {
						return nil, err
					}
					initiator := callEntry.Edges.Initiator
					if initiator == nil {
						return nil, fmt.Errorf("call log missing initiator relationship")
					}
					if initiator.ID != uid {
						roomEdge := callEntry.Edges.Room
						if roomEdge == nil {
							return nil, ErrForbidden
						}
						membership, err := r.ensureRoomMember(p.Context, roomEdge.ID, uid)
						if err != nil {
							return nil, err
						}
						if membership.Role == roommembership.RoleMember {
							return nil, ErrForbidden
						}
					}
					builder := r.Client.CallLog.UpdateOneID(id)
					if status, ok := p.Args["status"].(string); ok && status != "" {
						builder.SetStatus(calllog.Status(status))
					}
					if endedAt, ok := p.Args["endedAt"].(time.Time); ok {
						builder.SetEndedAt(endedAt)
					}
					if err := builder.Exec(p.Context); err != nil {
						return nil, err
					}
					return r.Client.CallLog.Query().
						Where(calllog.IDEQ(id)).
						WithInitiator().
						WithParticipants(func(q *ent.CallParticipantQuery) {
							q.WithParticipant()
						}).
						WithRoom().
						Only(p.Context)
				},
			},
			"deleteCallLog": &graphql.Field{
				Type: graphql.Boolean,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					id, err := decodeID(p.Args["id"])
					if err != nil {
						return nil, err
					}
					callEntry, err := r.Client.CallLog.Query().Where(calllog.IDEQ(id)).WithInitiator().WithRoom().Only(p.Context)
					if err != nil {
						return nil, err
					}
					initiator := callEntry.Edges.Initiator
					if initiator == nil {
						return nil, fmt.Errorf("call log missing initiator relationship")
					}
					if initiator.ID != uid {
						roomEdge := callEntry.Edges.Room
						if roomEdge == nil {
							return nil, ErrForbidden
						}
						membership, err := r.ensureRoomMember(p.Context, roomEdge.ID, uid)
						if err != nil {
							return nil, err
						}
						if membership.Role == roommembership.RoleMember {
							return nil, ErrForbidden
						}
					}
					return true, r.Client.CallLog.DeleteOneID(id).Exec(p.Context)
				},
			},
			"addCallParticipant": &graphql.Field{
				Type: r.callParticipantType(),
				Args: graphql.FieldConfigArgument{
					"callId":        &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"participantId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"role":          &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					callID, err := decodeID(p.Args["callId"])
					if err != nil {
						return nil, err
					}
					callEntry, err := r.Client.CallLog.Query().Where(calllog.IDEQ(callID)).WithInitiator().WithRoom().Only(p.Context)
					if err != nil {
						return nil, err
					}
					initiator := callEntry.Edges.Initiator
					if initiator == nil {
						return nil, fmt.Errorf("call log missing initiator relationship")
					}
					if initiator.ID != uid {
						roomEdge := callEntry.Edges.Room
						if roomEdge == nil {
							return nil, ErrForbidden
						}
						membership, err := r.ensureRoomMember(p.Context, roomEdge.ID, uid)
						if err != nil {
							return nil, err
						}
						if membership.Role == roommembership.RoleMember {
							return nil, ErrForbidden
						}
					}
					participantID, err := decodeID(p.Args["participantId"])
					if err != nil {
						return nil, err
					}
					participant := r.Client.CallParticipant.Create().
						SetCallID(callID).
						SetParticipantID(participantID)
					if role, ok := p.Args["role"].(string); ok && role != "" {
						participant.SetRole(callparticipant.Role(role))
					}
					if _, err := participant.Save(p.Context); err != nil {
						return nil, err
					}
					return r.Client.CallParticipant.Query().
						Where(callparticipant.HasCallWith(calllog.IDEQ(callID)), callparticipant.HasParticipantWith(user.ID(participantID))).
						WithCall(func(q *ent.CallLogQuery) {
							q.WithInitiator()
						}).
						WithParticipant().
						Only(p.Context)
				},
			},
			"updateCallParticipant": &graphql.Field{
				Type: r.callParticipantType(),
				Args: graphql.FieldConfigArgument{
					"callId":        &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"participantId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"role":          &graphql.ArgumentConfig{Type: graphql.String},
					"leftAt":        &graphql.ArgumentConfig{Type: graphql.DateTime},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					callID, err := decodeID(p.Args["callId"])
					if err != nil {
						return nil, err
					}
					callEntry, err := r.Client.CallLog.Query().Where(calllog.IDEQ(callID)).WithInitiator().WithRoom().Only(p.Context)
					if err != nil {
						return nil, err
					}
					initiator := callEntry.Edges.Initiator
					if initiator == nil {
						return nil, fmt.Errorf("call log missing initiator relationship")
					}
					if initiator.ID != uid {
						roomEdge := callEntry.Edges.Room
						if roomEdge == nil {
							return nil, ErrForbidden
						}
						membership, err := r.ensureRoomMember(p.Context, roomEdge.ID, uid)
						if err != nil {
							return nil, err
						}
						if membership.Role == roommembership.RoleMember {
							return nil, ErrForbidden
						}
					}
					participantID, err := decodeID(p.Args["participantId"])
					if err != nil {
						return nil, err
					}
					builder := r.Client.CallParticipant.Update().
						Where(callparticipant.HasCallWith(calllog.IDEQ(callID)), callparticipant.HasParticipantWith(user.ID(participantID)))
					if role, ok := p.Args["role"].(string); ok && role != "" {
						builder.SetRole(callparticipant.Role(role))
					}
					if leftAt, ok := p.Args["leftAt"].(time.Time); ok {
						builder.SetLeftAt(leftAt)
					}
					if _, err := builder.Save(p.Context); err != nil {
						return nil, err
					}
					return r.Client.CallParticipant.Query().
						Where(callparticipant.HasCallWith(calllog.IDEQ(callID)), callparticipant.HasParticipantWith(user.ID(participantID))).
						WithCall(func(q *ent.CallLogQuery) {
							q.WithInitiator()
						}).
						WithParticipant().
						Only(p.Context)
				},
			},
			"removeCallParticipant": &graphql.Field{
				Type: graphql.Boolean,
				Args: graphql.FieldConfigArgument{
					"callId":        &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"participantId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					callID, err := decodeID(p.Args["callId"])
					if err != nil {
						return nil, err
					}
					callEntry, err := r.Client.CallLog.Query().Where(calllog.IDEQ(callID)).WithInitiator().WithRoom().Only(p.Context)
					if err != nil {
						return nil, err
					}
					initiator := callEntry.Edges.Initiator
					if initiator == nil {
						return nil, fmt.Errorf("call log missing initiator relationship")
					}
					if initiator.ID != uid {
						roomEdge := callEntry.Edges.Room
						if roomEdge == nil {
							return nil, ErrForbidden
						}
						membership, err := r.ensureRoomMember(p.Context, roomEdge.ID, uid)
						if err != nil {
							return nil, err
						}
						if membership.Role == roommembership.RoleMember {
							return nil, ErrForbidden
						}
					}
					participantID, err := decodeID(p.Args["participantId"])
					if err != nil {
						return nil, err
					}
					_, err = r.Client.CallParticipant.Delete().
						Where(callparticipant.HasCallWith(calllog.IDEQ(callID)), callparticipant.HasParticipantWith(user.ID(participantID))).
						Exec(p.Context)
					if err != nil {
						return nil, err
					}
					return true, nil
				},
			},
		},
	}
}

func (r *Resolver) subscriptionFields() graphql.ObjectConfig {
	return graphql.ObjectConfig{
		Name: "Subscription",
		Fields: graphql.Fields{
			"notifications": &graphql.Field{
				Type: r.notificationType(),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source, nil
				},
				Subscribe: func(p graphql.ResolveParams) (interface{}, error) {
					uid, err := auth.UserIDFromContext(p.Context)
					if err != nil {
						return nil, ErrUnauthorized
					}
					ch, unsubscribe := r.notificationBroker.Subscribe(uid)
					stream := make(chan interface{})
					go func() {
						defer close(stream)
						defer unsubscribe()
						for {
							select {
							case <-p.Context.Done():
								return
							case notif, ok := <-ch:
								if !ok {
									return
								}
								if recipientIDFromNotification(notif) != uid {
									continue
								}
								select {
								case stream <- notif:
								case <-p.Context.Done():
									return
								}
							}
						}
					}()
					return stream, nil
				},
			},
		},
	}
}

func decodeID(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case string:
		var id int
		if _, err := fmt.Sscanf(v, "%d", &id); err != nil {
			return 0, fmt.Errorf("invalid id: %w", err)
		}
		return id, nil
	default:
		return 0, fmt.Errorf("unsupported id type: %T", value)
	}
}

func decodeIDList(value interface{}) []int {
	if value == nil {
		return nil
	}
	var raw []interface{}
	switch v := value.(type) {
	case []interface{}:
		raw = v
	case []int:
		return v
	case []string:
		result := make([]int, 0, len(v))
		for _, item := range v {
			id, err := decodeID(item)
			if err == nil {
				result = append(result, id)
			}
		}
		return result
	default:
		if err := mapstructure.Decode(v, &raw); err != nil {
			return nil
		}
	}
	result := make([]int, 0, len(raw))
	for _, item := range raw {
		if id, err := decodeID(item); err == nil {
			result = append(result, id)
		}
	}
	return result
}
