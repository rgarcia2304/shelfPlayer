BINARY=shelfplayer
INSTALL_PATH=$(HOME)/.local/bin/$(BINARY)

build:
	go build -o $(BINARY) .

install: build
	mkdir -p $(HOME)/.local/bin
	mv $(BINARY) $(INSTALL_PATH)
	@echo "installed to $(INSTALL_PATH)"

uninstall:
	rm -f $(INSTALL_PATH)
	@echo "uninstalled"

clean:
	rm -f $(BINARY)

.PHONY: build install uninstall clean
