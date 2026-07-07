APP_NAME := kindergarten-dienstplan
APP_ID := com.example.kindergarten-dienstplan
FYNE_CROSS := fyne-cross

.PHONY: build clean run tidy package-windows package-darwin package-linux package-all

build:
	go build -o $(APP_NAME) .

run: build
	./$(APP_NAME)

tidy:
	go mod tidy

package-windows:
	$(FYNE_CROSS) windows -engine podman -app-id $(APP_ID) -icon Icon.png

package-darwin:
	$(FYNE_CROSS) darwin -engine podman -app-id $(APP_ID) -icon Icon.png

package-linux:
	$(FYNE_CROSS) linux -engine podman -app-id $(APP_ID) -icon Icon.png

package-all: package-windows package-darwin package-linux

clean:
	rm -f $(APP_NAME)
	rm -rf fyne-cross
