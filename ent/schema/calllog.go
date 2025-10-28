package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// CallLog holds the schema definition for the CallLog entity.
type CallLog struct {
	ent.Schema
}

// Fields of the CallLog.
func (CallLog) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("status").Values("missed", "completed", "cancelled", "ongoing").Default("ongoing"),
		field.Time("started_at").Default(time.Now),
		field.Time("ended_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now),
	}
}

// Edges of the CallLog.
func (CallLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("initiator", User.Type).
			Unique().
			Required(),
		edge.To("room", Room.Type).
			Unique(),
		edge.From("participants", CallParticipant.Type).Ref("call"),
	}
}
