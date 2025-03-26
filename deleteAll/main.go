package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	livestream "cloud.google.com/go/video/livestream/apiv1"
	"cloud.google.com/go/video/livestream/apiv1/livestreampb"
	"google.golang.org/api/iterator"
)

var projectID string = "<PROJECT_NUMBER>"
var location string = "<REGION>"

// stopChannel stops a channel.
func stopChannel(ctx context.Context, client *livestream.Client, channelName string) error {
	stopReq := &livestreampb.StopChannelRequest{
		Name: channelName,
	}
	op, err := client.StopChannel(ctx, stopReq)
	if err != nil {
		return fmt.Errorf("StopChannel: %w", err)
	}
	// Corrected error handling for op.Wait()
	_, err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("StopChannel Wait: %w", err)
	}
	log.Printf("Stopped channel: %s\n", channelName)
	return nil
}

func listInputs(ctx context.Context, client *livestream.Client, projectID, location string) ([]*livestreampb.Input, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)
	inputIterator := client.ListInputs(ctx, &livestreampb.ListInputsRequest{
		Parent: parent,
	})
	var inputs []*livestreampb.Input
	for {
		input, err := inputIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("ListInputsIterator: %w", err)
		}
		inputs = append(inputs, input)
	}
	return inputs, nil
}

func listChannels(ctx context.Context, client *livestream.Client, projectID, location string) ([]*livestreampb.Channel, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)
	channelIterator := client.ListChannels(ctx, &livestreampb.ListChannelsRequest{
		Parent: parent,
	})
	var channels []*livestreampb.Channel
	for {
		channel, err := channelIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("ListChannelsIterator: %w", err)
		}
		channels = append(channels, channel)
	}
	return channels, nil
}

func listEvents(ctx context.Context, client *livestream.Client, projectID, location string, channelName string) ([]*livestreampb.Event, error) {
	parent := channelName // Parent is the channel name for events.
	eventIterator := client.ListEvents(ctx, &livestreampb.ListEventsRequest{
		Parent: parent,
	})
	var events []*livestreampb.Event
	for {
		event, err := eventIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("ListEventsIterator: %w", err)
		}
		events = append(events, event)
	}
	return events, nil
}

func deleteAllInputs(w io.Writer, projectID, location string) error {
	ctx := context.Background()
	client, err := livestream.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("NewClient: %w", err)
	}
	defer client.Close()

	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)

	// List all inputs
	inputIterator := client.ListInputs(ctx, &livestreampb.ListInputsRequest{
		Parent: parent,
	})

	// Delete each input
	for {
		input, err := inputIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("ListInputsIterator: %w", err)
		}

		// Delete the input
		inputName := input.GetName()
		log.Printf("Deleting input: %s\n", inputName)
		_, err = client.DeleteInput(ctx, &livestreampb.DeleteInputRequest{
			Name: inputName,
		})
		if err != nil {
			log.Printf("Error deleting input %s: %v\n", inputName, err) // Log and continue
			// Don't return here, continue deleting other inputs
		}
	}
	return nil
}

func deleteAllChannels(w io.Writer, projectID, location string) error {
	ctx := context.Background()
	client, err := livestream.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("NewClient: %w", err)
	}
	defer client.Close()

	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)

	// List all channels
	channelIterator := client.ListChannels(ctx, &livestreampb.ListChannelsRequest{
		Parent: parent,
	})

	// Stop and then Delete each channel
	for {
		channel, err := channelIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("ListChannelsIterator: %w", err)
		}

		channelName := channel.GetName()
		// Stop the channel first.
		err = stopChannel(ctx, client, channelName)
		if err != nil {
			log.Printf("Error stopping channel %s: %v\n", channelName, err)
			// keep going and try to delete other channels
		}

		// Delete the channel
		log.Printf("Deleting channel: %s\n", channelName)
		_, err = client.DeleteChannel(ctx, &livestreampb.DeleteChannelRequest{
			Name: channelName,
		})
		if err != nil {
			log.Printf("Error deleting channel %s: %v\n", channelName, err) // Log and continue
			// Don't return here, continue deleting other channels
		}
	}
	return nil
}

func main() {
	ctx := context.Background()
	client, err := livestream.NewClient(ctx)
	if err != nil {
		log.Printf("Error creating client: %v", err)
		return
	}
	defer client.Close()

	// List all inputs
	inputs, err := listInputs(ctx, client, projectID, location)
	if err != nil {
		log.Printf("Error listing inputs: %v", err)
		return
	}
	log.Printf("Inputs:\n")
	for _, input := range inputs {
		log.Printf("- %s\n", input.GetName())
	}

	// List all channels
	channels, err := listChannels(ctx, client, projectID, location)
	if err != nil {
		log.Printf("Error listing channels: %v", err)
		return
	}
	log.Printf("Channels:\n")
	for _, channel := range channels {
		log.Printf("- %s\n", channel.GetName())
		//list events for each channel
		events, err := listEvents(ctx, client, projectID, location, channel.GetName())
		if err != nil {
			log.Printf("Error listing events for channel %s: %v", channel.GetName(), err)
			return
		}
		if len(events) > 0 {
			log.Printf("  Events for channel %s:\n", channel.GetName())
			for _, event := range events {
				log.Printf("    - %s\n", event.GetName())
			}
		}

	}

	// Delete all channels first
	err = deleteAllChannels(os.Stdout, projectID, location)
	if err != nil {
		log.Printf("Error deleting channels: %v", err)
	}

	// Delete all inputs
	err = deleteAllInputs(os.Stdout, projectID, location)
	if err != nil {
		log.Printf("Error deleting inputs: %v", err)
	}
	log.Println("Finished deleting all channels and inputs.")
}
