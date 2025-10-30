package main

import (
	"context"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/adamtc007/KYC-DSL/api/pb"
	"github.com/adamtc007/KYC-DSL/internal/ui"
)

func main() {
	go run()
	app.Main()
}

func run() {
	// Create window
	w := app.NewWindow(
		app.Title("KYC-DSL CBU Graph Viewer"),
		app.Size(unit.Dp(1024), unit.Dp(768)),
	)

	// Create theme
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	var ops op.Ops

	// Connect to gRPC server
	serverAddr := os.Getenv("GRPC_SERVER")
	if serverAddr == "" {
		serverAddr = "localhost:50051"
	}

	log.Printf("Connecting to gRPC server at %s...", serverAddr)
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	// Create gRPC client
	client := pb.NewCbuGraphServiceClient(conn)

	// Fetch CBU graph
	cbuId := os.Getenv("CBU_ID")
	if cbuId == "" {
		cbuId = "BLACKROCK-GLOBAL"
	}

	log.Printf("Fetching CBU graph: %s...", cbuId)
	ctx := context.Background()
	graph, err := client.GetGraph(ctx, &pb.GetCbuRequest{CbuId: cbuId})
	if err != nil {
		log.Fatalf("Failed to fetch graph: %v", err)
	}

	log.Printf("Loaded graph: %s with %d entities and %d relationships",
		graph.Name, graph.EntityCount, graph.RelationshipCount)

	// Create graph view
	view := ui.NewGraphView()
	view.SetGraph(graph)

	// Event loop
	for {
		switch e := w.Event().(type) {
		case system.DestroyEvent:
			log.Println("Window closed")
			return

		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			
			// Draw graph
			view.Layout(gtx, th)

			// Submit frame
			e.Frame(gtx.Ops)
		}
	}
}
