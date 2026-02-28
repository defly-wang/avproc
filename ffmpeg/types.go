package ffmpeg

type MediaInfo struct {
	Filename    string
	Format      string
	Duration    string
	DurationSec float64
	Size        string
	Bitrate     string
	VideoTracks []VideoTrack
	AudioTracks []AudioTrack
}

type VideoTrack struct {
	Codec     string
	Width     int
	Height    int
	FrameRate string
	Bitrate   string
}

type AudioTrack struct {
	Codec      string
	SampleRate string
	Channels   int
	Bitrate    string
}

type Progress struct {
	Percent    float64
	Time       float64
	Speed      string
	OutputSize string
}
