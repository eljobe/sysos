libarchcode:
	$(MAKE) -C src

clean:
	$(MAKE) -C src clean

.PHONY: go_bindings
go_bindings:
	$(MAKE) -C src go_bindings
all:
	$(MAKE) -C src all
