package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// RoomMembership holds the schema definition for the RoomMembership entity.
type RoomMembership struct {
	ent.Schema
}

// Fields of the RoomMembership.
func (RoomMembership) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("role").Values("owner", "admin", "member").Default("member"),
		field.Bool("can_post").Default(true),
		field.Bool("can_call").Default(true),
		field.Time("joined_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the RoomMembership.
func (RoomMembership) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Unique().
			Required(),
		edge.To("room", Room.Type).
			Unique().
			Required(),
	}
}

// Indexes of the RoomMembership.
func (RoomMembership) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("user", "room").Unique(),
	}
}
