.PHONY: all install

all: install

install:
	@./bin/install.sh $(TARGET_DIR)
