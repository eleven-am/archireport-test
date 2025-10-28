package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Media holds the schema definition for the Media entity.
type Media struct {
	ent.Schema
}

// Fields of the Media.
func (Media) Fields() []ent.Field {
	return []ent.Field{
		field.String("filename").NotEmpty(),
		field.String("content_type").NotEmpty(),
		field.String("storage_path").NotEmpty(),
		field.String("checksum").NotEmpty(),
		field.Int64("size_bytes").NonNegative(),
		field.Time("created_at").Default(time.Now),
	}
}

// Edges of the Media.
func (Media) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("uploader", User.Type).
			Unique().
			Required(),
		edge.To("message", Message.Type).
			Unique(),
	}
}
