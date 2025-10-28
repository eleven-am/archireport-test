package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Message holds the schema definition for the Message entity.
type Message struct {
	ent.Schema
}

// Fields of the Message.
func (Message) Fields() []ent.Field {
	return []ent.Field{
		field.String("cipher_text").NotEmpty(),
		field.String("content_type").Default("text/plain"),
		field.String("encryption_scheme").Default("signal"),
		field.Bool("edited").Default(false),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Message.
func (Message) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("sender", User.Type).
			Unique().
			Required(),
		edge.To("room", Room.Type).
			Unique().
			Required(),
		edge.From("media", Media.Type).Ref("message"),
	}
}
