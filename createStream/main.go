package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	livestream "cloud.google.com/go/video/livestream/apiv1"
	"cloud.google.com/go/video/livestream/apiv1/livestreampb"
	"google.golang.org/protobuf/encoding/protojson"
)

var runningNumber string = "01"
var projectID string = "<PROJECT_NUMBER>"
var location string = "<REGION>"
var channelID string = "livestream-channel-" + runningNumber
var inputID string = "livestream-input-" + runningNumber
var gcsoutput string = "gs://<GCP_STORAGE_BUCKET>/" + inputID

// createInputIfNotExists creates an input endpoint if it does not exist.
func createInputIfNotExists(w io.Writer, projectID, location, inputID string) error {
	ctx := context.Background()
	client, err := livestream.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("NewClient: %w", err)
	}
	defer client.Close()

	inputName := fmt.Sprintf("projects/%s/locations/%s/inputs/%s", projectID, location, inputID)
	reqGet := &livestreampb.GetInputRequest{
		Name: inputName,
	}

	existingInput, err := client.GetInput(ctx, reqGet)
	if err == nil {
		log.Printf("Input %s already exists.\n", inputName)
		inputInfo := map[string]interface{}{
			"inputID": inputName,
			"uri":     existingInput.GetUri(),
		}
		prettyJSON, err := json.MarshalIndent(inputInfo, "", "  ")
		if err != nil {
			return fmt.Errorf("JSON marshal error: %w", err)
		}

		w.Write(prettyJSON)
		w.Write([]byte("\n"))

		//write input json to file
		inputFile, err := os.Create(inputID + ".json")
		if err != nil {
			return fmt.Errorf("Error creating input file: %v", err)
		}
		defer inputFile.Close()
		_, err = inputFile.Write(prettyJSON)
		if err != nil {
			return fmt.Errorf("Error writing to input file: %v", err)
		}
		return nil
	}

	reqCreate := &livestreampb.CreateInputRequest{
		Parent:  fmt.Sprintf("projects/%s/locations/%s", projectID, location),
		InputId: inputID,
		Input: &livestreampb.Input{
			Type: livestreampb.Input_RTMP_PUSH,
		},
	}

	op, err := client.CreateInput(ctx, reqCreate)
	if err != nil {
		return fmt.Errorf("CreateInput: %w", err)
	}
	response, err := op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Wait: %w", err)
	}

	log.Printf("Input created: %s\n", response.Name)
	inputInfo := map[string]interface{}{
		"inputID": response.Name,
		"uri":     response.GetUri(),
	}
	prettyJSON, err := json.MarshalIndent(inputInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON marshal error: %w", err)
	}

	w.Write(prettyJSON)
	w.Write([]byte("\n"))

	//write input json to file
	inputFile, err := os.Create(inputID + ".json")
	if err != nil {
		log.Printf("Error creating input file: %v", err)
		return nil
	}
	defer inputFile.Close()
	_, err = inputFile.Write(prettyJSON)
	if err != nil {
		return fmt.Errorf("Error writing to input file: %v", err)
	}
	return nil
}

func createChannelIfNotExists(w io.Writer, projectID, location, channelID, requestJSONPath string) error {
	ctx := context.Background()
	client, err := livestream.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("NewClient: %w", err)
	}
	defer client.Close()

	channelName := fmt.Sprintf("projects/%s/locations/%s/channels/%s", projectID, location, channelID)
	reqGet := &livestreampb.GetChannelRequest{
		Name: channelName,
	}

	existingChannel, err := client.GetChannel(ctx, reqGet)

	if err == nil {
		log.Printf("Channel %s already exists.\n", channelName)
		channelInfo := map[string]interface{}{
			"channelID": channelName,
			"inputID":   fmt.Sprintf("projects/%s/locations/%s/inputs/%s", projectID, location, inputID),
			"gcsoutput": existingChannel.GetOutput().GetUri(),
		}

		prettyJSON, err := json.MarshalIndent(channelInfo, "", "  ")
		if err != nil {
			return fmt.Errorf("JSON marshal error: %w", err)
		}

		w.Write(prettyJSON)
		w.Write([]byte("\n"))

		//write channel json to file
		channelFile, err := os.Create(channelID + ".json")
		if err != nil {
			return fmt.Errorf("Error creating channel file: %v", err)
		}
		defer channelFile.Close()
		_, err = channelFile.Write(prettyJSON)
		if err != nil {
			return fmt.Errorf("Error writing channel file: %v", err)
		}

		return nil
	}

	requestJSON, err := os.ReadFile(requestJSONPath)
	requestStr := string(requestJSON)
	requestStr = strings.ReplaceAll(requestStr, "<GCS_OUTPUT>", gcsoutput)
	requestStr = strings.ReplaceAll(requestStr, "<GCP_OTHER_INFO>", fmt.Sprintf("projects/%s/locations/%s/inputs/%s", projectID, location, inputID))

	if err != nil {
		return fmt.Errorf("ReadFile: %w", err)
	}

	channel := &livestreampb.Channel{}
	err = protojson.Unmarshal([]byte(requestStr), channel)
	if err != nil {
		return fmt.Errorf("Unmarshal: %w", err)
	}

	req := &livestreampb.CreateChannelRequest{
		Parent:    fmt.Sprintf("projects/%s/locations/%s", projectID, location),
		ChannelId: channelID,
		Channel:   channel,
	}

	op, err := client.CreateChannel(ctx, req)
	if err != nil {
		return fmt.Errorf("CreateChannel: %w", err)
	}

	response, err := op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Wait: %w", err)
	}

	log.Printf("Channel created: %s\n", response.Name)

	channelInfo := map[string]interface{}{
		"channelID": response.Name,
		"inputID":   fmt.Sprintf("projects/%s/locations/%s/inputs/%s", projectID, location, inputID),
		"gcsoutput": response.GetOutput().GetUri(),
	}

	prettyJSON, err := json.MarshalIndent(channelInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON marshal error: %w", err)
	}

	w.Write(prettyJSON)
	w.Write([]byte("\n"))

	//write channel json to file
	channelFile, err := os.Create(channelID + ".json")
	if err != nil {
		log.Printf("Error creating channel file: %v", err)
		return nil
	}
	defer channelFile.Close()
	_, err = channelFile.Write(prettyJSON)
	if err != nil {
		return fmt.Errorf("Error writing channel file: %v", err)
	}
	return nil
}

func getChannelState(client *livestream.Client, channelName string) (livestreampb.Channel_StreamingState, error) {
	ctx := context.Background()
	req := &livestreampb.GetChannelRequest{
		Name: channelName,
	}
	channel, err := client.GetChannel(ctx, req)
	if err != nil {
		return livestreampb.Channel_STREAMING_STATE_UNSPECIFIED, fmt.Errorf("GetChannel: %w", err)
	}
	return channel.StreamingState, nil
}

func printChannelState(client *livestream.Client, channelName string) {
	state, err := getChannelState(client, channelName)
	if err != nil {
		log.Printf("Error getting channel state: %v", err)
		return
	}
	log.Printf("Channel %s streaming state: %s\n", channelName, state.String())
}

func main() {
	err := createInputIfNotExists(os.Stdout, projectID, location, inputID)
	if err != nil {
		log.Printf("Error during input creation: %v", err)
	}
	channelRequest := "request.json"
	err = createChannelIfNotExists(os.Stdout, projectID, location, channelID, channelRequest)
	if err != nil {
		log.Printf("Error creating channel: %v", err)
	}

	ctx := context.Background()
	client, err := livestream.NewClient(ctx)
	if err != nil {
		log.Printf("Error creating client: %v", err)
		return
	}
	defer client.Close()

	channelName := fmt.Sprintf("projects/%s/locations/%s/channels/%s", projectID, location, channelID)

	// Get the current channel state
	currentState, err := getChannelState(client, channelName)
	if err != nil {
		log.Printf("Error getting initial channel state: %v", err)
		return
	}

	// Start the channel only if it's not already started.
	if currentState != livestreampb.Channel_STREAMING && currentState != livestreampb.Channel_AWAITING_INPUT {
		reqStart := &livestreampb.StartChannelRequest{
			Name: channelName,
		}

		opStart, err := client.StartChannel(ctx, reqStart)
		if err != nil {
			log.Printf("Error starting channel: %v", err)
			return
		}

		_, err = opStart.Wait(ctx)
		if err != nil {
			log.Printf("Error waiting for start operation: %v", err)
			return
		}
		log.Printf("Channel %s started.\n", channelName)
	} else {
		log.Printf("Channel %s already started or starting, skipping start operation.\n", channelName)
	}

	// Continuously display the channel's streaming state.
	for {
		currentState, err = getChannelState(client, channelName)
		if err != nil {
			log.Printf("Error getting channel state: %v", err)
			// Don't return here, continue to try and get the state.
		}
		printChannelState(client, channelName)
		time.Sleep(5 * time.Second)
	}
}
