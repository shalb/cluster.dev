package profiler

type Profiler struct {
	timeLines map[string]*TimeLine
}

func NewProfiler() Profiler {
	return Profiler{
		timeLines: map[string]*TimeLine{
			"main": NewTimeLine(),
		},
	}
}

var Global = NewProfiler()

func (f *Profiler) MainTimeLine() *TimeLine {
	return f.timeLines["main"]
}

func (f *Profiler) CustomTimeLine(name string) *TimeLine {
	res, exists := f.timeLines[name]
	if !exists {
		t := NewTimeLine()
		res = t
		f.timeLines[name] = t
	}
	return res
}
