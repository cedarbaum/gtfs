package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/jamespfennell/gtfs"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "GTFS parser",
		Usage: "parse GTFS static and realtime feeds",
		Commands: []*cli.Command{
			{
				Name:  "static",
				Usage: "parse a GTFS static message",
				Action: func(*cli.Context) error {
					fmt.Print("static")
					return nil
				},
			},
			{
				Name:  "realtime",
				Usage: "parse a GTFS realtime message",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "print additional data about each trip and vehicle",
					},
					&cli.BoolFlag{
						Name:  "nyct",
						Usage: "use the New York City Transit GTFS realtime extension",
					},
				},
				ArgsUsage: "path",
				Action: func(ctx *cli.Context) error {
					args := ctx.Args()
					if args.Len() == 0 {
						return fmt.Errorf("a path to the GTFS realtime message was not provided")
					}
					path := ctx.Args().First()
					b, err := os.ReadFile(path)
					if err != nil {
						return fmt.Errorf("failed to read file %s: %w", path, err)
					}

					opts := gtfs.ParseRealtimeOptions{}
					if ctx.Bool("nyct") {
						opts.UseNyctExtension = true
						americaNewYorkTimezone, err := time.LoadLocation("America/New_York")
						if err == nil {
							opts.Timezone = americaNewYorkTimezone
						}
					}
					realtime, err := gtfs.ParseRealtime(b, &opts)
					if err != nil {
						return fmt.Errorf("failed to parse message: %w", err)
					}
					fmt.Printf("%d trips:\n", len(realtime.Trips))
					for _, trip := range realtime.Trips {
						fmt.Printf("- %s\n", formatTrip(trip, 2, ctx.Bool("verbose")))
					}
					return nil
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func formatTrip(trip gtfs.Trip, indent int, printStopTimes bool) string {
	var b strings.Builder
	tc := color.New(color.FgCyan)
	vc := color.New(color.FgMagenta)
	sc := color.New(color.FgGreen)
	newLine := fmt.Sprintf("\n%*s", indent, "")
	fmt.Fprintf(&b,
		"TripID %s  RouteID %s  DirectionID %s  StartDate %s  StartTime %s%s",
		tc.Sprint(trip.ID.ID),
		tc.Sprint(trip.ID.RouteID),
		tc.Sprint(trip.ID.DirectionID),
		tc.Sprint(trip.ID.StartDate.Format("2006-01-02")),
		tc.Sprint(trip.ID.StartTime),
		newLine,
	)
	if trip.Vehicle != nil {
		fmt.Fprintf(&b, "Vehicle: ID %s%s", vc.Sprint(trip.Vehicle.GetID().ID), newLine)
	} else {
		fmt.Fprintf(&b, "Vehicle: <none>%s", newLine)
	}

	if printStopTimes {
		fmt.Fprintf(&b, "Stop times (%d):%s", len(trip.StopTimeUpdates), newLine)
		for _, stopTime := range trip.StopTimeUpdates {
			fmt.Fprintf(&b,
				"  StopSeq %s  StopID %s  Arrival %s  Departure %s  NyctTrack %s%s",
				sc.Sprint(unPtrI(stopTime.StopSequence)),
				sc.Sprint(unPtr(stopTime.StopID)),
				unPtrT(stopTime.GetArrival().Time, sc),
				unPtrT(stopTime.GetDeparture().Time, sc),
				sc.Sprint(unPtr(stopTime.NyctTrack)),
				newLine,
			)
		}
	} else {
		fmt.Fprintf(&b, "Num stop times: %d (show with -v)%s", len(trip.StopTimeUpdates), newLine)
	}

	return b.String()
}

func unPtr(s *string) string {
	if s == nil {
		return "<none>"
	}
	return *s
}

func unPtrI(s *uint32) string {
	if s == nil {
		return "<none>"
	}
	return fmt.Sprintf("%d", s)
}

func unPtrT(t *time.Time, c *color.Color) string {
	if t == nil {
		return "<none>"
	}
	return c.Sprint(t.String())
}
