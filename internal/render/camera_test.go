package render

import (
	"math"
	"testing"
)

func TestCameraDoesNotFollowCloseRangeMidpoint(t *testing.T) {
	for _, side := range []float32{-1, 1} {
		for step := 0; step <= 120; step++ {
			walker := side * (-1.5 + float32(step)*.018)
			if got := cameraCenter(0, 210, walker, side*1.5); got != 0 {
				t.Fatalf("approach pulled the camera and stationary fighter: %v", got)
			}
		}
	}
}

func TestCameraPansOnlyEnoughToKeepFightersVisible(t *testing.T) {
	for _, side := range []float32{-1, 1} {
		center := cameraCenter(0, 210, side*2.5, side*3.6)
		if center*side <= 0 || math.Abs(float64(center)) > .66 {
			t.Fatalf("unexpected camera movement at edge: %v", center)
		}
		if got := cameraCenter(center, 210, side*2, side*3); got != center {
			t.Fatal("camera recenters when fighters return inside safe area")
		}
	}
}
