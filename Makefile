WHISPER_DIR := $(abspath ./whisper.cpp)
MODELS_DIR := $(abspath ./models)
MODEL_NAME := base.en
MODEL_FILE_NAME := ggml-$(MODEL_NAME).bin
# TODO make this configurable
INCLUDE_PATH := $(WHISPER_DIR)
LIBRARY_PATH := $(WHISPER_DIR)

backend := wcpp

whisper-cpp-pre-reqs:
ifeq ($(wildcard $(WHISPER_DIR)/*),)
	@echo "fetching whisper repo"
	@${MAKE} -C ../ fetch-whisper
endif

ifeq ($(wildcard $(WHISPER_DIR)/libwhisper.a),)
	@echo "building whisper lib"
	@${MAKE} -C ../ build-whisper-lib
endif

ifeq ($(wildcard $(MODELS_DIR)/$(MODEL_FILE_NAME)),)
	@echo "fetching model"
	@${MAKE} -C ../ fetch-model
endif

run-whisper-cpp: whisper-cpp-pre-reqs
	@echo "running client with whisper.cpp backend..."
	@C_INCLUDE_PATH=${INCLUDE_PATH} LIBRARY_PATH=${LIBRARY_PATH}  go run bot.go

build-whisper-cpp: 
	@echo "running client with whisper.cpp backend..."
	@C_INCLUDE_PATH=${INCLUDE_PATH} LIBRARY_PATH=${LIBRARY_PATH}  go build ${BUILD_FLAGS} -o ./

debug: whisper-cpp-pre-reqs
	@C_INCLUDE_PATH=${INCLUDE_PATH} LIBRARY_PATH=${LIBRARY_PATH} PKG_CONFIG_PATH= go run cmd/whisper.cpp/main.go --debug=true


build-whisper-lib:
	@${MAKE} -C ./whisper.cpp libwhisper.a

fetch-model:
	@${MAKE} -C ./whisper.cpp base.en
	@cp $(WHISPER_DIR)/models/ggml-$(MODEL_NAME).bin $(MODELS_DIR)

gen-api:
	goctl api go -api bot.api -dir ./ -style go_zero
	@echo "Generate API files successfully"