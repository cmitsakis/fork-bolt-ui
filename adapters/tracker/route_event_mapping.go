package tracker

import (
	"encoding/json"
	"time"

	"github.com/boreq/errors"
	"github.com/boreq/velo/domain"
	"github.com/boreq/velo/internal/eventsourcing"
)

var routeEventMapping = eventsourcing.Mapping{
	"RouteCreated_v1": eventsourcing.EventMapping{
		Marshal: func(event eventsourcing.Event) ([]byte, error) {
			e := event.(domain.RouteCreated)

			transportEvent := routeCreated{
				UUID: e.UUID.String(),
			}

			for _, p := range e.Points {
				transportEvent.Points = append(
					transportEvent.Points,
					point{
						Time: p.Time(),
						Position: position{
							Latitude:  p.Position().Latitude().Float64(),
							Longitude: p.Position().Longitude().Float64(),
						},
						Altitude: p.Altitude().Float64(),
					},
				)
			}

			return json.Marshal(transportEvent)
		},
		Unmarshal: func(bytes []byte) (eventsourcing.Event, error) {
			var transportEvent routeCreated

			if err := json.Unmarshal(bytes, &transportEvent); err != nil {
				return nil, errors.Wrap(err, "could not unmarshal json")
			}

			uuid, err := domain.NewRouteUUID(transportEvent.UUID)
			if err != nil {
				return nil, errors.Wrap(err, "could not create a uuid")
			}

			var points []domain.Point

			for _, p := range transportEvent.Points {
				latitude, err := domain.NewLatitude(p.Position.Latitude)
				if err != nil {
					return nil, errors.Wrap(err, "could not create a latitude")
				}

				longitude, err := domain.NewLongitude(p.Position.Longitude)
				if err != nil {
					return nil, errors.Wrap(err, "could not create a longitude")
				}

				position := domain.NewPosition(latitude, longitude)

				altitude := domain.NewAltitude(p.Altitude)

				point, err := domain.NewPoint(p.Time, position, altitude)
				if err != nil {
					return nil, errors.Wrap(err, "could not create a point")
				}

				points = append(points, point)
			}

			return domain.RouteCreated{
				UUID:   uuid,
				Points: points,
			}, nil
		},
	},
}

type routeCreated struct {
	UUID   string `json:"uuid"`
	Points []point
}

type point struct {
	Time     time.Time
	Position position
	Altitude float64
}

type position struct {
	Latitude  float64
	Longitude float64
}
