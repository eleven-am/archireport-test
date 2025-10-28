package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").NotEmpty().Unique(),
		field.String("display_name").NotEmpty(),
		field.String("email").NotEmpty().Unique(),
		field.String("avatar_url").Optional().Nillable(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("last_seen_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("memberships", RoomMembership.Type).Ref("user"),
		edge.From("messages", Message.Type).Ref("sender"),
		edge.From("uploaded_media", Media.Type).Ref("uploader"),
		edge.From("owned_rooms", Room.Type).Ref("owner"),
		edge.From("contacts", Contact.Type).Ref("owner"),
		edge.From("contact_entries", Contact.Type).Ref("contact"),
		edge.From("favourites", Favourite.Type).Ref("user"),
		edge.From("initiated_calls", CallLog.Type).Ref("initiator"),
		edge.From("call_participations", CallParticipant.Type).Ref("participant"),
	}
}
