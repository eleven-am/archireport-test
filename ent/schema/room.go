package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Room holds the schema definition for the Room entity.
type Room struct {
	ent.Schema
}

// Fields of the Room.
func (Room) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("description").Default(""),
		field.Bool("is_private").Default(false),
		field.Bool("is_direct").Default(false),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Room.
func (Room) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("owner", User.Type).
			Unique().
			Required(),
		edge.From("memberships", RoomMembership.Type).Ref("room"),
		edge.From("messages", Message.Type).Ref("room"),
		edge.From("favourites", Favourite.Type).Ref("room"),
		edge.From("call_logs", CallLog.Type).Ref("room"),
	}
}
