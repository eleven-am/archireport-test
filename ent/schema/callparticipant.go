package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// CallParticipant holds the schema definition for the CallParticipant entity.
type CallParticipant struct {
	ent.Schema
}

// Fields of the CallParticipant.
func (CallParticipant) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("role").Values("caller", "callee", "observer").Default("callee"),
		field.Time("joined_at").Default(time.Now),
		field.Time("left_at").Optional().Nillable(),
	}
}

// Edges of the CallParticipant.
func (CallParticipant) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("call", CallLog.Type).
			Unique().
			Required(),
		edge.To("participant", User.Type).
			Unique().
			Required(),
	}
}
