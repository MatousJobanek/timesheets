APP_NAME := kindergarten-dienstplan
APP_ID := com.example.kindergarten-dienstplan
FYNE_CROSS_FLAGS := -engine podman -app-id $(APP_ID) -icon Icon.png -env GOTOOLCHAIN=auto

.PHONY: build clean run tidy fmt package-windows package-darwin package-linux package-all

fmt:
	gofmt -w .
	go mod tidy

build:
	go build -o $(APP_NAME) .

run: build
	./$(APP_NAME)

tidy:
	go mod tidy

package-windows:
	fyne-cross windows $(FYNE_CROSS_FLAGS)

package-darwin:
	fyne-cross darwin $(FYNE_CROSS_FLAGS)

package-linux:
	fyne-cross linux $(FYNE_CROSS_FLAGS)

package-all: package-windows package-darwin package-linux

clean:
	rm -f $(APP_NAME)
	rm -rf fyne-cross
