// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/vertexai/genai"
)

var prompt = "What does this picture look like? Provide a short answer in less than 8 words."

func guess(ctx context.Context, jpegData []byte) (genai.Part, error) {
	client, err := genai.NewClient(ctx, projectID, location)
	if err != nil {
		return nil, fmt.Errorf("unable to create client: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel(modelName)
	model.Temperature = 0.4

	// Retry policy (max 3), because we sometimes get
	// "rpc error: code = InvalidArgument desc = Request contains an invalid argument"
	for attempt := 1; attempt <= 3; attempt++ {
		img := genai.ImageData("jpeg", jpegData)
		res, err := model.GenerateContent(ctx, img, genai.Text(prompt))
		if err != nil {
			log.Printf("Error calling GenerateContent: %v", err)
			if attempt == 3 {
				log.Println("Giving up on guess after 3 attempts to call GenerateContent")
				return nil, err
			}
			log.Println("Attempt", attempt, "failed, retrying GenerateContent")
			time.Sleep(500 * time.Millisecond)
			continue
		}
		answer := res.Candidates[0].Content.Parts[0]
		return answer, nil
	}
	panic("unreachable")
}
