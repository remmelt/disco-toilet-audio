img := "remmelt/disco-toilet-audio:latest"

.PHONY: run
run:
	@VOLUME=2 \
    HUE_USERNAME=qz4MvxEkwXmyJrVLkncH5arHKfBux0-keoOf4l9A \
    HUE_BRIDGE_IP=192.168.178.126 \
    MPD_IP=192.168.178.89 \
    PLAYLIST=WC \
    DAY_START=7:30 \
    DAY_END=22:00 \
	go run main.go

.PHONY: clean
clean:
	go clean


.PHONY: push
push: clean
	docker buildx build --push --platform linux/amd64 --tag ${img} .
