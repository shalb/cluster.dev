package profiler

import "time"

type TimePoint struct {
	name          string
	start         time.Time
	previousPoint *TimePoint
	nextPoint     *TimePoint
}

type TimeLine struct {
	points     []*TimePoint
	startPoint *TimePoint
}

func (p *TimePoint) Duration() time.Duration {
	if p.previousPoint == nil {
		return 0
	}
	return p.start.Sub(p.previousPoint.start)
}

func (p *TimeLine) Duration() time.Duration {
	if len(p.points) == 0 {
		return 0
	}
	return p.points[len(p.points)-1].start.Sub(p.startPoint.start)
}

func (t *TimeLine) Start() {
	t.startPoint = &TimePoint{
		name:          "start",
		nextPoint:     nil,
		previousPoint: nil,
		start:         time.Now(),
	}
	t.points = make([]*TimePoint, 0)
}

func (t *TimeLine) SetPoint(name string) {
	if t.startPoint == nil {
		t.Start()
	}
	now := time.Now()
	p := &TimePoint{
		previousPoint: nil,
		name:          name,
		start:         now,
	}
	if len(t.points) > 0 {
		p.previousPoint = t.points[len(t.points)-1]
		p.previousPoint.nextPoint = p

	} else {
		t.startPoint.nextPoint = p
	}
	t.points = append(t.points, p)
}

func (t *TimeLine) IsEmpty() bool {
	return len(t.points) == 0
}

func NewTimeLine() *TimeLine {
	return &TimeLine{
		points: []*TimePoint{},
	}
}
