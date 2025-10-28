package schema

import (
        "time"

        "entgo.io/ent"
        "entgo.io/ent/schema/edge"
        "entgo.io/ent/schema/field"
)

// Notification holds the schema definition for the Notification entity.
type Notification struct {
        ent.Schema
}

// Fields of the Notification.
func (Notification) Fields() []ent.Field {
        return []ent.Field{
                field.String("kind").NotEmpty(),
                field.String("cipher_text").NotEmpty(),
                field.String("encryption_scheme").Default("signal"),
                field.Bool("read").Default(false),
                field.Time("created_at").Default(time.Now),
                field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
        }
}

// Edges of the Notification.
func (Notification) Edges() []ent.Edge {
        return []ent.Edge{
                edge.To("recipient", User.Type).
                        Unique().
                        Required(),
                edge.To("room", Room.Type).
                        Unique(),
                edge.To("message", Message.Type).
                        Unique(),
        }
}
