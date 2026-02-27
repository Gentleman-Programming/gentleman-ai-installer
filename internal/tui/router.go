package tui

type Route struct {
	Forward  Screen
	Backward Screen
}

var linearRoutes = map[Screen]Route{
	ScreenWelcome:        {Forward: ScreenDetection},
	ScreenDetection:      {Forward: ScreenAgents, Backward: ScreenWelcome},
	ScreenAgents:         {Forward: ScreenPersona, Backward: ScreenDetection},
	ScreenPersona:        {Forward: ScreenPreset, Backward: ScreenAgents},
	ScreenPreset:         {Forward: ScreenDependencyTree, Backward: ScreenPersona},
	ScreenDependencyTree: {Forward: ScreenReview, Backward: ScreenPreset},
	ScreenReview:         {Forward: ScreenInstalling, Backward: ScreenDependencyTree},
	ScreenInstalling:     {Forward: ScreenComplete, Backward: ScreenReview},
	ScreenComplete:       {Backward: ScreenInstalling},
	ScreenBackups:        {Backward: ScreenWelcome},
}

func NextScreen(screen Screen) (Screen, bool) {
	route, ok := linearRoutes[screen]
	if !ok || route.Forward == ScreenUnknown {
		return ScreenUnknown, false
	}

	return route.Forward, true
}

func PreviousScreen(screen Screen) (Screen, bool) {
	route, ok := linearRoutes[screen]
	if !ok || route.Backward == ScreenUnknown {
		return ScreenUnknown, false
	}

	return route.Backward, true
}
