package logger

func TestLogLevel() {
	Wrapper.LogError("Error testing %s %d", "error", 1)
	Wrapper.LogFatal("Fatal testing %s %d", "fatal", 2)
	Wrapper.LogInfo("Error testing %s %d", "info", 3)
	Wrapper.LogDebug("Fatal testing %s %d", "debug", 4)
	Wrapper.LogTrace("Fatal testing %s %d", "trace", 5)
}
