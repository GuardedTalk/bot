package engine

import (
	"fmt"
	"math"
	"sync"
)

// FIXME make these configurable
const (
	// This is determined by the hyperparameter configuration that whisper was trained on.
	// See more here: https://github.com/ggerganov/whisper.cpp/issues/909
	SampleRate   = 16000 // 16kHz
	sampleRateMs = SampleRate / 1000
	// This determines how much audio we will be passing to whisper inference.
	// We will buffer up to (whisperSampleWindowMs - pcmSampleRateMs) of old audio and then add
	// audioSampleRateMs of new audio onto the end of the buffer for inference
	sampleWindowMs = 24000 // 24 second sample window
	windowSize     = sampleWindowMs * sampleRateMs
	// This is the minimum ammount of audio we want to buffer before running inference
	// 2 seconds of audio samples
	windowMinSize = 2000 * sampleRateMs
	// This determines how often we will try to run inference.
	// We will buffer (pcmSampleRateMs * whisperSampleRate / 1000) samples and then run inference
	pcmSampleRateMs = 500 // FIXME PLEASE MAKE ME AN CONFIG PARAM
	pcmWindowSize   = pcmSampleRateMs * sampleRateMs

	// this is an arbitrary number I picked after testing a bit
	// feel free to play around
	energyThresh  = 0.0005
	silenceThresh = 0.015
)

type Engine struct {
	sync.Mutex
	// Buffer to store new audio. When this fills up we will try to run inference
	pcmWindow []float32
	// Buffer to store old and new audio to run inference on.
	// By inferring on old and new audio we can help smooth out cross word boundaries
	window               []float32
	lastHandledTimestamp uint32
	wp                   *WhisperModel
	isSpeaking           bool
}

type Transcriber interface {
	Transcribe(audioData []float32) (Transcription, error)
}

type TranscriptionSegment struct {
	StartTimestamp uint32 `json:"startTimestamp"`
	EndTimestamp   uint32 `json:"endTimestamp"`
	Text           string `json:"text"`
}

type Transcription struct {
	From           uint32
	Transcriptions []TranscriptionSegment
}

func New() (*Engine, error) {
	wp, err := NewWhisperModel("./models/ggml-base.en.bin")
	if err != nil {
		panic(err)
	}
	return &Engine{
		window:               make([]float32, 0, windowSize),
		pcmWindow:            make([]float32, 0, pcmWindowSize),
		lastHandledTimestamp: 0,
		wp:                   wp,
		isSpeaking:           false,
	}, nil
}

func (e *Engine) Write(pcm []float32, timestamp uint32) {
	e.writeVAD(pcm, timestamp)
}

// XXX DANGER XXX
// This is highly experiemential and will probably crash in very interesting ways. I have deadlines
// and am hacking towards what I want to demo. Use at your own risk :D
// XXX DANGER XXX
//
// writeVAD only buffers audio if somone is speaking. It will run inference after the audio transitions from
// speaking to not speaking
func (e *Engine) writeVAD(pcm []float32, timestamp uint32) {
	// TODO normalize PCM and see if we can make it better
	// endTimestamp is the latest packet timestamp + len of the audio in the packet
	// FIXME make these timestamps make sense
	e.Lock()
	defer e.Unlock()
	if len(e.pcmWindow)+len(pcm) > pcmWindowSize {
		// This shouldn't happen hopefully...
		// Logger.Infof("GOING TO OVERFLOW PCM WINDOW BY %d", len(e.pcmWindow)+len(pcm)-pcmWindowSize)
	}
	e.pcmWindow = append(e.pcmWindow, pcm...)
	if len(e.pcmWindow) >= pcmWindowSize {
		// reset window
		defer func() {
			e.pcmWindow = e.pcmWindow[:0]
		}()

		isSpeaking := VAD(e.pcmWindow)

		defer func() {
			e.isSpeaking = isSpeaking
		}()

		if isSpeaking && e.isSpeaking {
			fmt.Println("STILL SPEAKING")
			// Logger.Debug("STILL SPEAKING")
			// add to buffer and wait
			// FIXME make sure we have space
			e.window = append(e.window, e.pcmWindow...)
			return
		} else if isSpeaking && !e.isSpeaking {
			fmt.Println("JUST STARTED SPEAKING")
			e.isSpeaking = isSpeaking
			// we just started speaking, add to buffer and wait
			// FIXME make sure we have space
			e.window = append(e.window, e.pcmWindow...)
			return
		} else if !isSpeaking && e.isSpeaking {
			fmt.Println("JUST STOPPED SPEAKING")
			e.window = append(e.window, e.pcmWindow...)
			transcript, err := e.wp.Transcribe(e.window)
			if err != nil {
				fmt.Println("----------->", err)
			}
			fmt.Println("----------->", transcript.Transcriptions)

		} else if !isSpeaking && !e.isSpeaking {
			if len(e.window) != 0 {
				fmt.Printf("running whisper inference with %d window length \n", len(e.window))
				fmt.Println("NOT SPEAKING")
				return
			}
		}
	}
}

func VAD(frame []float32) bool {
	// Compute frame energy
	energy := float32(0)
	for i := 0; i < len(frame); i++ {
		energy += frame[i] * frame[i]
	}
	energy /= float32(len(frame))

	// Apply energy threshold
	if energy < energyThresh {
		return false
	}

	// Compute frame silence
	silence := float32(0)
	for i := 0; i < len(frame); i++ {
		silence += float32(math.Abs(float64(frame[i])))
	}
	silence /= float32(len(frame))

	// Apply silence threshold
	if silence < silenceThresh {
		return false
	}

	return true
}
