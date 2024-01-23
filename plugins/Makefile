SOURCES := example/example.go kafka/kafka.go
BINARIES := $(SOURCES:.go=.so)
OUTPUT_DIR := plugin

build: $(BINARIES)

%.so: %.go
	go build -o $@ -buildmode=plugin $<
	mv $@ .
	
clean:
	rm -f $(BINARIES)

