package graphql

import (
	"context"
	"time"

	"github.com/graphql-go/graphql"

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
)

func (r *Resolver) userType() *graphql.Object {
	if r.userObj == nil {
		r.userObj = graphql.NewObject(graphql.ObjectConfig{
			Name: "User",
			Fields: graphql.FieldsThunk(func() graphql.Fields {
				return graphql.Fields{
					"id":          &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
					"username":    &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: resolveStringField("Username")},
					"displayName": &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: resolveStringField("DisplayName")},
					"email":       &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: resolveStringField("Email")},
					"avatarUrl":   &graphql.Field{Type: graphql.String, Resolve: resolveStringPointerField("AvatarURL")},
					"createdAt":   &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime), Resolve: resolveTimeField("CreatedAt")},
					"updatedAt":   &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime), Resolve: resolveTimeField("UpdatedAt")},
					"lastSeenAt":  &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime), Resolve: resolveTimeField("LastSeenAt")},
					"memberships": &graphql.Field{
						Type: graphql.NewList(r.roomMembershipType()),
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							usr := p.Source.(*ent.User)
							return r.Client.RoomMembership.Query().
								Where(roommembership.HasUserWith(user.IDEQ(usr.ID))).
								WithRoom().
								All(p.Context)
						},
					},
					"notifications": &graphql.Field{
						Type: graphql.NewList(r.notificationType()),
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							usr := p.Source.(*ent.User)
							return r.Client.Notification.Query().
								Where(notification.HasRecipientWith(user.IDEQ(usr.ID))).
								Order(ent.Desc(notification.FieldCreatedAt)).
								All(p.Context)
						},
					},
				}
			}),
		})
	}
	return r.userObj
}

func (r *Resolver) roomType() *graphql.Object {
	if r.roomObj == nil {
		r.roomObj = graphql.NewObject(graphql.ObjectConfig{
			Name: "Room",
			Fields: graphql.FieldsThunk(func() graphql.Fields {
				return graphql.Fields{
					"id":          &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
					"name":        &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: resolveStringField("Name")},
					"description": &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: resolveStringField("Description")},
					"isPrivate":   &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean), Resolve: resolveBoolField("IsPrivate")},
					"isDirect":    &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean), Resolve: resolveBoolField("IsDirect")},
					"createdAt":   &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime), Resolve: resolveTimeField("CreatedAt")},
					"updatedAt":   &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime), Resolve: resolveTimeField("UpdatedAt")},
					"owner": &graphql.Field{
						Type: r.userType(),
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							rm := p.Source.(*ent.Room)
							if rm.Edges.Owner != nil {
								return rm.Edges.Owner, nil
							}
							return r.Client.Room.Query().Where(room.IDEQ(rm.ID)).WithOwner().Only(p.Context)
						},
					},
				}
			}),
		})
	}
	return r.roomObj
}

func (r *Resolver) roomMembershipType() *graphql.Object {
	if r.roomMembershipObj == nil {
		r.roomMembershipObj = graphql.NewObject(graphql.ObjectConfig{
			Name: "RoomMembership",
			Fields: graphql.FieldsThunk(func() graphql.Fields {
				return graphql.Fields{
					"id":        &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
					"role":      &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: resolveEnumField()},
					"canPost":   &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean), Resolve: resolveBoolField("CanPost")},
					"canCall":   &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean), Resolve: resolveBoolField("CanCall")},
					"joinedAt":  &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime), Resolve: resolveTimeField("JoinedAt")},
					"updatedAt": &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime), Resolve: resolveTimeField("UpdatedAt")},
					"user": &graphql.Field{
						Type: r.userType(),
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							membership := p.Source.(*ent.RoomMembership)
							if membership.Edges.User != nil {
								return membership.Edges.User, nil
							}
							return membership.QueryUser().Only(p.Context)
						},
					},
					"room": &graphql.Field{
						Type: r.roomType(),
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							membership := p.Source.(*ent.RoomMembership)
							if membership.Edges.Room != nil {
								return membership.Edges.Room, nil
							}
							return membership.QueryRoom().Only(p.Context)
						},
					},
				}
			}),
		})
	}
	return r.roomMembershipObj
}

func (r *Resolver) messageType() *graphql.Object {
	if r.messageObj == nil {
		r.messageObj = graphql.NewObject(graphql.ObjectConfig{
			Name: "Message",
			Fields: graphql.FieldsThunk(func() graphql.Fields {
				return graphql.Fields{
					"id":               &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
					"cipherText":       &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: resolveStringField("CipherText")},
					"contentType":      &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: resolveStringField("ContentType")},
					"encryptionScheme": &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: resolveStringField("EncryptionScheme")},
					"edited":           &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean), Resolve: resolveBoolField("Edited")},
					"createdAt":        &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime), Resolve: resolveTimeField("CreatedAt")},
					"updatedAt":        &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime), Resolve: resolveTimeField("UpdatedAt")},
					"sender": &graphql.Field{
						Type: r.userType(),
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							msg := p.Source.(*ent.Message)
							if msg.Edges.Sender != nil {
								return msg.Edges.Sender, nil
							}
							return r.Client.Message.Query().Where(message.IDEQ(msg.ID)).WithSender().Only(p.Context)
						},
					},
					"room": &graphql.Field{
						Type: r.roomType(),
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							msg := p.Source.(*ent.Message)
							if msg.Edges.Room != nil {
								return msg.Edges.Room, nil
							}
							return r.Client.Message.Query().Where(message.IDEQ(msg.ID)).WithRoom().Only(p.Context)
						},
					},
					"media": &graphql.Field{
						Type: graphql.NewList(r.mediaType()),
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							msg := p.Source.(*ent.Message)
							return r.Client.Media.Query().
								Where(media.HasMessageWith(message.IDEQ(msg.ID))).
								All(p.Context)
						},
					},
				}
			}),
		})
	}
	return r.messageObj
}

func (r *Resolver) mediaType() *graphql.Object {
	if r.mediaObj == nil {
		r.mediaObj = graphql.NewObject(graphql.ObjectConfig{
			Name: "Media",
			Fields: graphql.FieldsThunk(func() graphql.Fields {
				return graphql.Fields{
					"id":          &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
					"filename":    &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: resolveStringField("Filename")},
					"contentType": &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: resolveStringField("ContentType")},
					"storagePath": &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: resolveStringField("StoragePath")},
					"checksum":    &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: resolveStringField("Checksum")},
					"sizeBytes":   &graphql.Field{Type: graphql.NewNonNull(graphql.Int), Resolve: resolveInt64Field("SizeBytes")},
					"createdAt":   &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime), Resolve: resolveTimeField("CreatedAt")},
					"uploader": &graphql.Field{
						Type: r.userType(),
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							item := p.Source.(*ent.Media)
							if item.Edges.Uploader != nil {
								return item.Edges.Uploader, nil
							}
							return r.Client.Media.Query().Where(media.IDEQ(item.ID)).WithUploader().Only(p.Context)
						},
					},
				}
			}),
		})
	}
	return r.mediaObj
}

func (r *Resolver) notificationType() *graphql.Object {
	if r.notificationObj == nil {
		r.notificationObj = graphql.NewObject(graphql.ObjectConfig{
			Name: "Notification",
			Fields: graphql.FieldsThunk(func() graphql.Fields {
				return graphql.Fields{
					"id":               &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
					"kind":             &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: resolveStringField("Kind")},
					"cipherText":       &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: resolveStringField("CipherText")},
					"encryptionScheme": &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: resolveStringField("EncryptionScheme")},
					"read":             &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean), Resolve: resolveBoolField("Read")},
					"createdAt":        &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime), Resolve: resolveTimeField("CreatedAt")},
					"updatedAt":        &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime), Resolve: resolveTimeField("UpdatedAt")},
					"recipient": &graphql.Field{
						Type: r.userType(),
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							notification := p.Source.(*ent.Notification)
							if notification.Edges.Recipient != nil {
								return notification.Edges.Recipient, nil
							}
							return notification.QueryRecipient().Only(p.Context)
						},
					},
					"room": &graphql.Field{
						Type: r.roomType(),
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							notification := p.Source.(*ent.Notification)
							if notification.Edges.Room != nil {
								return notification.Edges.Room, nil
							}
							rm, err := notification.QueryRoom().Only(p.Context)
							if ent.IsNotFound(err) {
								return nil, nil
							}
							return rm, err
						},
					},
					"message": &graphql.Field{
						Type: r.messageType(),
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							notification := p.Source.(*ent.Notification)
							if notification.Edges.Message != nil {
								return notification.Edges.Message, nil
							}
							msg, err := notification.QueryMessage().Only(p.Context)
							if ent.IsNotFound(err) {
								return nil, nil
							}
							return msg, err
						},
					},
				}
			}),
		})
	}
	return r.notificationObj
}

func (r *Resolver) contactType() *graphql.Object {
	if r.contactObj == nil {
		r.contactObj = graphql.NewObject(graphql.ObjectConfig{
			Name: "Contact",
			Fields: graphql.FieldsThunk(func() graphql.Fields {
				return graphql.Fields{
					"id":          &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
					"alias":       &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: resolveStringField("Alias")},
					"isFavourite": &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean), Resolve: resolveBoolField("IsFavourite")},
					"isBlocked":   &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean), Resolve: resolveBoolField("IsBlocked")},
					"createdAt":   &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime), Resolve: resolveTimeField("CreatedAt")},
					"updatedAt":   &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime), Resolve: resolveTimeField("UpdatedAt")},
					"owner": &graphql.Field{
						Type: r.userType(),
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							entry := p.Source.(*ent.Contact)
							if entry.Edges.Owner != nil {
								return entry.Edges.Owner, nil
							}
							return r.Client.Contact.Query().Where(contact.IDEQ(entry.ID)).WithOwner().Only(p.Context)
						},
					},
					"contact": &graphql.Field{
						Type: r.userType(),
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							entry := p.Source.(*ent.Contact)
							if entry.Edges.Contact != nil {
								return entry.Edges.Contact, nil
							}
							return r.Client.Contact.Query().Where(contact.IDEQ(entry.ID)).WithContact().Only(p.Context)
						},
					},
				}
			}),
		})
	}
	return r.contactObj
}

func (r *Resolver) favouriteType() *graphql.Object {
	if r.favouriteObj == nil {
		r.favouriteObj = graphql.NewObject(graphql.ObjectConfig{
			Name: "Favourite",
			Fields: graphql.FieldsThunk(func() graphql.Fields {
				return graphql.Fields{
					"id":        &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
					"createdAt": &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime), Resolve: resolveTimeField("CreatedAt")},
					"user":      &graphql.Field{Type: r.userType(), Resolve: resolveEdgeUser(r)},
					"room":      &graphql.Field{Type: r.roomType(), Resolve: resolveEdgeRoom(r)},
				}
			}),
		})
	}
	return r.favouriteObj
}

func (r *Resolver) callLogType() *graphql.Object {
	if r.callLogObj == nil {
		r.callLogObj = graphql.NewObject(graphql.ObjectConfig{
			Name: "CallLog",
			Fields: graphql.FieldsThunk(func() graphql.Fields {
				return graphql.Fields{
					"id":        &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
					"status":    &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: resolveEnumField()},
					"startedAt": &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime), Resolve: resolveTimeField("StartedAt")},
					"endedAt":   &graphql.Field{Type: graphql.DateTime, Resolve: resolveOptionalTimeField("EndedAt")},
					"createdAt": &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime), Resolve: resolveTimeField("CreatedAt")},
					"initiator": &graphql.Field{Type: r.userType(), Resolve: resolveEdgeInitiator(r)},
					"room":      &graphql.Field{Type: r.roomType(), Resolve: resolveEdgeRoom(r)},
					"participants": &graphql.Field{
						Type: graphql.NewList(r.callParticipantType()),
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							call := p.Source.(*ent.CallLog)
							return r.Client.CallParticipant.Query().
								Where(callparticipant.HasCallWith(calllog.IDEQ(call.ID))).
								WithParticipant().
								All(p.Context)
						},
					},
				}
			}),
		})
	}
	return r.callLogObj
}

func (r *Resolver) callParticipantType() *graphql.Object {
	if r.callParticipantObj == nil {
		r.callParticipantObj = graphql.NewObject(graphql.ObjectConfig{
			Name: "CallParticipant",
			Fields: graphql.FieldsThunk(func() graphql.Fields {
				return graphql.Fields{
					"id":          &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
					"role":        &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: resolveEnumField()},
					"joinedAt":    &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime), Resolve: resolveTimeField("JoinedAt")},
					"leftAt":      &graphql.Field{Type: graphql.DateTime, Resolve: resolveOptionalTimeField("LeftAt")},
					"participant": &graphql.Field{Type: r.userType(), Resolve: resolveEdgeParticipant(r)},
				}
			}),
		})
	}
	return r.callParticipantObj
}

func resolveStringField(field string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		return getField(p.Source, field)
	}
}

func resolveStringPointerField(field string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		if value, err := getField(p.Source, field); err == nil {
			switch v := value.(type) {
			case *string:
				if v != nil {
					return *v, nil
				}
				return nil, nil
			case string:
				if v == "" {
					return nil, nil
				}
				return v, nil
			}
			return value, nil
		} else {
			return nil, err
		}
	}
}

func resolveTimeField(field string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		value, err := getField(p.Source, field)
		if err != nil {
			return nil, err
		}
		if t, ok := value.(time.Time); ok {
			return t, nil
		}
		return nil, nil
	}
}

func resolveOptionalTimeField(field string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		value, err := getField(p.Source, field)
		if err != nil {
			return nil, err
		}
		switch v := value.(type) {
		case *time.Time:
			if v != nil {
				return *v, nil
			}
			return nil, nil
		case time.Time:
			if v.IsZero() {
				return nil, nil
			}
			return v, nil
		default:
			return nil, nil
		}
	}
}

func resolveBoolField(field string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		value, err := getField(p.Source, field)
		if err != nil {
			return nil, err
		}
		if b, ok := value.(bool); ok {
			return b, nil
		}
		return nil, nil
	}
}

func resolveInt64Field(field string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		value, err := getField(p.Source, field)
		if err != nil {
			return nil, err
		}
		switch v := value.(type) {
		case int64:
			return v, nil
		case int:
			return int64(v), nil
		}
		return nil, nil
	}
}

func resolveEnumField() graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		switch v := p.Source.(type) {
		case *ent.RoomMembership:
			return v.Role, nil
		case *ent.CallParticipant:
			return v.Role, nil
		case *ent.CallLog:
			return v.Status, nil
		default:
			value, err := getField(p.Source, "Role")
			if err != nil {
				return getField(p.Source, "Status")
			}
			return value, nil
		}
	}
}

func resolveEdgeUser(r *Resolver) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		switch v := p.Source.(type) {
		case *ent.Favourite:
			if v.Edges.User != nil {
				return v.Edges.User, nil
			}
			return r.Client.Favourite.Query().Where(favourite.IDEQ(v.ID)).WithUser().Only(p.Context)
		default:
			return nil, nil
		}
	}
}

func resolveEdgeRoom(r *Resolver) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		switch v := p.Source.(type) {
		case *ent.Favourite:
			if v.Edges.Room != nil {
				return v.Edges.Room, nil
			}
			return r.Client.Favourite.Query().Where(favourite.IDEQ(v.ID)).QueryRoom().Only(p.Context)
		case *ent.CallLog:
			if v.Edges.Room != nil {
				return v.Edges.Room, nil
			}
			roomEntity, err := v.QueryRoom().Only(p.Context)
			if err != nil {
				if ent.IsNotFound(err) {
					return nil, nil
				}
				return nil, err
			}
			return roomEntity, nil
		default:
			return nil, nil
		}
	}
}

func resolveEdgeInitiator(r *Resolver) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		call := p.Source.(*ent.CallLog)
		if call.Edges.Initiator != nil {
			return call.Edges.Initiator, nil
		}
		return r.Client.CallLog.Query().Where(calllog.IDEQ(call.ID)).WithInitiator().Only(p.Context)
	}
}

func resolveEdgeParticipant(r *Resolver) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		participant := p.Source.(*ent.CallParticipant)
		if participant.Edges.Participant != nil {
			return participant.Edges.Participant, nil
		}
		return r.Client.CallParticipant.Query().
			Where(callparticipant.IDEQ(participant.ID)).
			WithParticipant().
			Only(p.Context)
	}
}

func getField(source interface{}, name string) (interface{}, error) {
	if source == nil {
		return nil, nil
	}
	switch v := source.(type) {
	case map[string]interface{}:
		return v[name], nil
	default:
		return graphql.DefaultResolveFn(graphql.ResolveParams{Context: context.Background(), Source: source, Info: graphql.ResolveInfo{FieldName: name}})
	}
}
