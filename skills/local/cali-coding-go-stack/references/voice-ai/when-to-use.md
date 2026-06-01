# Voice AI: When to Use LiveKit + Gemini

## Architecture

```
┌─────────────┐     WebRTC      ┌─────────────┐    SIP/gRPC    ┌─────────────┐
│   Browser   │ ◄──────────────►│   LiveKit   │ ◄─────────────►│   Gemini    │
│  (mic/cam)  │                 │    Server   │                │  Live API   │
└─────────────┘                 └─────────────┘                └─────────────┘
```

## When to USE

- **Interactive voice agents** - chatbots that speak
- **Real-time transcription** - meetings, calls
- **Visual assistants** - screen sharing + voice
- **Automated IVR/NPS** - phone support
- **Remote education** - tutoring with voice

## When NOT to USE

- **Simple chat** - text solves it (use goai SDK)
- **Text generation/embeddings** - no voice needed
- **Non-interactive audio analysis** - batch processing
- **No microphone available** - accessibility

## LiveKit (WebRTC)

**What it is:** WebRTC platform (real-time video/audio)

**Characteristics:**
- Multi-participants (rooms)
- Session recording
- Screen sharing
- Built-in TURN/STUN
- SDKs for web, mobile, server

**When to choose LiveKit alone:**
- When you need video + audio
- When multiple users interact
- When you need to record sessions
- For conference calls

```go
// Server-side LiveKit (without Gemini)
room, err := lksdk.CreateRoom(&lksdk.RoomConfig{
    Name:        "my-room",
    EmptyTimeout: 5 * 60, // 5 min
})

participant, err := room.Join(token, &lksdkParticipant{
    Name: "user-123",
})
```

## Gemini Live API

**What it is:** Real-time conversation API with Gemini (video + audio)

**Characteristics:**
- Bidirectional audio (input + output)
- Optional video (camera)
- Function calling during conversation
- Context-aware (remembers conversation)
- Multimodal (sees what's on screen)

**When to choose Gemini Live API:**
- Voice assistants / chatbots
- AI that needs to "see" (screen sharing)
- Human-AI interaction by voice
- Asynchronous tasks with voice feedback

```go
// Gemini Live API (server-side)
geminiService.Connect(ctx, roomName, token)

// Client starts with user gesture (REQUIRED for autoplay)
await page.evaluate(() => {
    window.enableMicrophone(); // only called after user click
});
```

## LiveKit + Gemini Integration

### Key: User Gesture for Audio

Browsers **require** user interaction to enable microphone. Always:

```javascript
// ❌ WRONG - will fail due to autoplay
micButton.addEventListener('mouseenter', () => {
    room.localParticipant.enableMicrophone();
});

// ✅ CORRECT - only after click
micButton.addEventListener('click', () => {
    room.localParticipant.enableMicrophone();
});
```

### Storage Polyfill for LiveKit

LiveKit uses `localStorage`. In environments without it (incognito, strict):
```javascript
// polyfill for localStorage
if (!window.localStorage) {
    window.localStorage = {
        data: {},
        getItem: function(key) { return this.data[key] || null; },
        setItem: function(key, value) { this.data[key] = value; },
        removeItem: function(key) { delete this.data[key]; }
    };
}
```

## Configuration

### Required environment variables:

```env
LIVEKIT_URL=wss://your-livekit-server.com
LIVEKIT_API_KEY=your-api-key
LIVEKIT_API_SECRET=your-secret
GEMINI_API_KEY=your-gemini-key
```

### Dependencies:

```go
// go.mod
github.com/livekit/livekit-server v1.5.0  // server
github.com/livekit/livekit-sdk v1.5.0       // client (if using SDK)
github.com/google/generative-ai-go v0.14.0  // Gemini
```

## Alternatives

| Need | Tool |
|------|------|
| Text + chat only | goai SDK |
| Audio synth (TTS) | ElevenLabs, OpenAI Audio |
| STT (speech-to-text) | Whisper, Google Speech |
| Video conferencing | Daily.co, Zoom, Jitsi |
| Conversational AI | Gemini Live, GPT-4o, Claude Voice |
