package engine

import (
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/media"
)

const (
	sampleRate  = 16000 // (16000)
	channels    = 1     // decode into 1 channel since that is what whisper.cpp wants
	frameSizeMs = 20
)

var frameSize = channels * frameSizeMs * sampleRate / 1000

// AudioEngine is used to convert RTP Opus packets to raw PCM audio to be sent to Whisper
// and to convert raw PCM audio from Coqui back to RTP Opus packets to be sent back over WebRTC
type AudioEngine struct {
	// RTP Opus packets to be converted to PCM
	rtpIn chan *rtp.Packet
	// RTP Opus packets converted from PCM to be sent over WebRTC
	mediaOut chan media.Sample

	dec *OpusDecoder
	enc *OpusEncoder
	// slice to hold raw pcm data during decoding
	pcm []float32
	// slice to hold binary encoded pcm data
	buf []byte

	sttEngine *Engine

	firstTimeStamp uint32
	// shouldInfer determines if we should run TTS inference or not
	shouldInfer atomic.Bool
}

func NewAudioEngine() (*AudioEngine, error) {
	dec, err := NewOpusDecoder(sampleRate, channels)
	if err != nil {
		return nil, err
	}

	// we use 2 channels for the output
	enc, err := NewOpusEncoder(2, frameSizeMs)
	if err != nil {
		return nil, err
	}

	var shouldInfer atomic.Bool
	shouldInfer.Store(true)
	stt, _ := New()
	ae := &AudioEngine{
		rtpIn:          make(chan *rtp.Packet),
		mediaOut:       make(chan media.Sample),
		pcm:            make([]float32, frameSize),
		buf:            make([]byte, frameSize*2),
		dec:            dec,
		enc:            enc,
		firstTimeStamp: 0,
		sttEngine:      stt,
		shouldInfer:    shouldInfer,
	}

	return ae, nil
}

func (a *AudioEngine) RtpIn() chan<- *rtp.Packet {
	return a.rtpIn
}

func (a *AudioEngine) MediaOut() <-chan media.Sample {
	return a.mediaOut
}

func (a *AudioEngine) Start() {
	fmt.Println("Starting audio engine")
	go a.decode()
}

// Pause stops the text to speech inference and simply drops incoming packets
func (a *AudioEngine) Pause() {
	fmt.Println("Pausing tts")
	a.shouldInfer.Swap(false)
}

// Unpause restarts the text to speech inference
func (a *AudioEngine) Unpause() {
	fmt.Println("Unpausing tts")
	a.shouldInfer.Swap(true)
}

// Encode takes in raw f32le pcm, encodes it into opus RTP packets and sends those over the rtpOut chan
func (a *AudioEngine) Encode(pcm []float32, inputChannelCount, inputSampleRate int) error {
	opusFrames, err := a.enc.Encode(pcm, inputChannelCount, inputSampleRate)
	if err != nil {
		fmt.Println(err, "error encoding pcm")
	}

	go a.sendMedia(opusFrames)

	return nil
}

// sendMedia turns opus frames into media samples and sends them on the channel
func (a *AudioEngine) sendMedia(frames []OpusFrame) {
	for _, f := range frames {
		sample := convertOpusToSample(f)
		a.mediaOut <- sample
		// this is important to properly pace the samples
		time.Sleep(time.Millisecond * 20)
	}

	// start inferring audio again
	a.Unpause()
}

func convertOpusToSample(frame OpusFrame) media.Sample {
	return media.Sample{
		Data:               frame.Data,
		PrevDroppedPackets: 0, // FIXME support dropping packets
		Duration:           time.Millisecond * 20,
	}
}

// decode reads over the in channel in a loop, decodes the RTP packets to raw PCM and sends the data on another channel
func (a *AudioEngine) decode() {
	for {
		pkt, ok := <-a.rtpIn
		if !ok {
			fmt.Println("rtpIn channel closed...")
			return
		}
		if !a.shouldInfer.Load() {
			continue
		}
		if a.firstTimeStamp == 0 {
			fmt.Println("Resetting timestamp bc firstTimeStamp is 0...  ", pkt.Timestamp)
			a.firstTimeStamp = pkt.Timestamp
		}

		if _, err := a.decodePacket(pkt); err != nil {
			fmt.Println(err, "error decoding opus packet ")
		}
	}
}

func (a *AudioEngine) decodePacket(pkt *rtp.Packet) (int, error) {
	_, err := a.dec.Decode(pkt.Payload, a.pcm)
	// we decode to float32 here since that is what whisper.cpp takes
	if err != nil {
		fmt.Println(err, "error decoding fb packet")
		return 0, err
	} else {
		timestampMS := (pkt.Timestamp - a.firstTimeStamp) / ((sampleRate / 1000) * 3)
		lengthOfRecording := uint32(len(a.pcm) / (sampleRate / 1000))
		timestampRecordingEnds := timestampMS + lengthOfRecording
		a.sttEngine.Write(a.pcm, timestampRecordingEnds)
		return convertToBytes(a.pcm, a.buf), nil
	}
}

func (a *AudioEngine) DecodePacket(pkt *rtp.Packet) (int, error) {
	return a.decodePacket(pkt)
}

// This function converts f32le to s16le bytes for writing to a file
func convertToBytes(in []float32, out []byte) int {
	currIndex := 0
	for i := range in {
		res := int16(math.Floor(float64(in[i] * 32767)))

		out[currIndex] = byte(res & 0b11111111)
		currIndex++

		out[currIndex] = (byte(res >> 8))
		currIndex++
	}
	return currIndex
}
