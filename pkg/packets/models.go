package packets

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Packet used to hold packet data
type Packet struct {
	ID         primitive.ObjectID `json:"id" bson:"id"`
	Device     string             `json:"device" bson:"Device"`
	Type       float64            `json:"type" bson:"Type"`
	SrcIP      string             `json:"src_ip" bson:"SrcIP"`
	DstIP      string             `json:"dst_ip" bson:"DstIP"`
	SrcPort    string             `json:"src_port" bson:"SrcPort"`
	DstPort    string             `json:"dst_port" bson:"DstPort"`
	Payload    string             `json:"payload" bson:"Payload"`
	CapturedAt *time.Time         `json:"captured_at" bson:"CapturedAt"`
}

// GetType returns a string representing it packet's type
func (p *Packet) GetType() string {
	switch p.Type {
	case 0:
		return "TCP"
	case 1:
		return "UDP"
	default:
		return "Unrecognized"
	}
}
