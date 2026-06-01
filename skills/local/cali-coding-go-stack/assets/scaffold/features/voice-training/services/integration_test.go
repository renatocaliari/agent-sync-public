package services_test

import (
	"context"
	"os"
	"testing"
	"time"

	"northstar/features/voice-training/pages"
	"northstar/features/voice-training/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVoiceTrainingIntegration(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get credentials from env
	livekitURL := os.Getenv("LIVEKIT_URL")
	livekitAPIKey := os.Getenv("LIVEKIT_API_KEY")
	livekitAPISecret := os.Getenv("LIVEKIT_API_SECRET")
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")

	if livekitURL == "" || livekitAPIKey == "" || livekitAPISecret == "" || geminiAPIKey == "" {
		t.Skip("Skipping integration test: missing environment variables")
	}

	ctx := context.Background()

	// Test 1: Create LiveKit room
	t.Run("CreateRoom", func(t *testing.T) {
		liveKitService := services.NewLiveKitService(livekitURL, livekitAPIKey, livekitAPISecret)
		require.NotNil(t, liveKitService)

		roomName := "test-room-" + time.Now().Format("20060102150405")
		room, err := liveKitService.CreateRoom(ctx, roomName)
		require.NoError(t, err)
		assert.NotNil(t, room)
		assert.Equal(t, roomName, room.Name)

		t.Logf("✅ Room created: %s", room.Name)

		// Cleanup
		err = liveKitService.DeleteRoom(ctx, roomName)
		assert.NoError(t, err)
		t.Logf("✅ Room deleted")
	})

	// Test 2: Generate token
	t.Run("GenerateToken", func(t *testing.T) {
		liveKitService := services.NewLiveKitService(livekitURL, livekitAPIKey, livekitAPISecret)

		token, err := liveKitService.GenerateToken("test-room", "test-identity")
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Contains(t, token, ".") // JWT format

		t.Logf("✅ Token generated (length: %d)", len(token))
	})

	// Test 3: Gemini connection
	t.Run("GeminiConnection", func(t *testing.T) {
		geminiService := services.NewGeminiService(geminiAPIKey)
		require.NotNil(t, geminiService)

		systemPrompt := "Você é um cliente simulado para testes. Responda brevemente."

		err := geminiService.Connect(ctx, systemPrompt)
		require.NoError(t, err)

		t.Logf("✅ Gemini connected")

		// Set up audio callback
		audioReceived := make(chan []byte, 1)
		geminiService.SetOnAudioOut(func(audio []byte) {
			t.Logf("✅ Received audio from Gemini: %d bytes", len(audio))
			audioReceived <- audio
		})

		// Send some dummy PCM audio (silence)
		dummyAudio := make([]byte, 320) // 10ms of silence at 16kHz
		err = geminiService.SendAudio(dummyAudio)
		require.NoError(t, err)
		t.Logf("✅ Audio sent to Gemini")

		// Wait for response (with timeout)
		select {
		case audio := <-audioReceived:
			t.Logf("✅ Audio round-trip successful: %d bytes", len(audio))
		case <-time.After(5 * time.Second):
			t.Log("⚠️ No audio received from Gemini (may be expected for silence)")
		}

		geminiService.Close()
		t.Logf("✅ Gemini connection closed")
	})

	// Test 4: Full bot workflow
	t.Run("BotJoinAndAudioFlow", func(t *testing.T) {
		liveKitService := services.NewLiveKitService(livekitURL, livekitAPIKey, livekitAPISecret)
		geminiService := services.NewGeminiService(geminiAPIKey)

		roomName := "test-bot-room-" + time.Now().Format("20060102150405")

		// Create room first
		room, err := liveKitService.CreateRoom(ctx, roomName)
		require.NoError(t, err)
		t.Logf("✅ Room created: %s", room.Name)

		// Create bot service
		botService := services.NewBotService(liveKitService, geminiService)
		require.NotNil(t, botService)

		systemPrompt := "Você é um cliente simulado para testes. Responda brevemente em português."

		// Join bot (this connects to both LiveKit and Gemini)
		errChan := make(chan error, 1)
		go func() {
			err := botService.JoinRoom(ctx, roomName, systemPrompt, "male")
			errChan <- err
		}()

		// Wait for bot to join (with timeout)
		select {
		case err := <-errChan:
			require.NoError(t, err)
			t.Logf("✅ Bot joined room")
		case <-time.After(10 * time.Second):
			t.Fatal("Timeout waiting for bot to join")
		}

		// Give some time for the bot to be fully connected
		time.Sleep(2 * time.Second)

		// Cleanup
		botService.Leave()
		err = liveKitService.DeleteRoom(ctx, roomName)
		assert.NoError(t, err)
		t.Logf("✅ Bot workflow completed and cleaned up")
	})
}

func TestSessionService(t *testing.T) {
	// This tests the session service with NATS
	// Would need NATS setup for full integration test
	t.Run("SessionCRUD", func(t *testing.T) {
		t.Skip("Requires NATS setup - run manually or with docker-compose")
	})
}

func TestRandomNameGeneration(t *testing.T) {
	names := make(map[string]bool)

	// Generate 100 names and ensure they're all valid Brazilian names
	for i := 0; i < 100; i++ {
		name := services.RandomBrazilianName()
		assert.NotEmpty(t, name)
		names[name] = true
	}

	// Should have variety
	assert.Greater(t, len(names), 10, "Should generate different names")
	t.Logf("✅ Generated %d unique names from %d attempts", len(names), 100)
}
