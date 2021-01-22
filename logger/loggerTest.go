package logger

func main() {
	LogError("Error testing %s %d", "error", 1)
	LogFatal("Fatal testing %s %d", "fatal", 2)
	LogInfo("Error testing %s %d", "info", 3)
	LogDebug("Fatal testing %s %d", "debug", 4)
	LogTrace("Fatal testing %s %d", "trace", 5)
}
