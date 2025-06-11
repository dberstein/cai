run:
	@GEMINI_API_KEY=$(shell pass Api/Gemini) PAGER="" go run main/main.go