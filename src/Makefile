VERSION := 0.1.0
OUT_DIR ?= $(CURDIR)/../build
BASE_DIR := $(CURDIR)/..
OBJ_DIR := $(OUT_DIR)/obj
LIB_DIR := $(OUT_DIR)/lib
BIN_DIR := $(OUT_DIR)/bin

# Ensure these directories exist
$(shell mkdir -p $(OBJ_DIR))
$(shell mkdir -p $(LIB_DIR))

ASFLAGS += -g -fpic
CFLAGS +=  -g -Wall -Werror
CLANG_ASFLAGS = -fno-integrated-as
LDFLAGS += -L .

###############################################################################
# Platform Configuration
###############################################################################
# Platform detection.
ifndef OS
    OS := $(shell uname -s)
endif

ifeq ($(OS),Windows_NT)
	PLATFORM = Windows
else
	ifeq ($(OS),Darwin)
		PLATFORM = Darwin
	else
		PLATFORM = Linux
	endif
endif

# ARM architecture detection
ifdef CC
  ARM = $(shell $(CC) -dM -E - < /dev/null | grep "aarch" | awk '{ print $$3 }')
  ifneq ($(findstring mingw, $(CC)),)
	ifneq ($(ARM),1)
		PLATFORM = Windows
	endif
  endif
else
  ARCH = $(shell uname -m)
  ARM = $(shell echo $(ARCH) | grep -E '^(arm|aarch64)' >/dev/null && echo 1 || echo 0)
endif

# Cross-platform compiler selection
# check for default, skip setting if user passed in specific cross-compilation lib
ifeq ($(CC), cc)
	ifeq ($(PLATFORM),Darwin)
		CC = clang
	else
		CC = gcc
	endif
endif

ifeq ($(CC),clang)
	ifneq ($(ARM),1)
		ASFLAGS += $(CLANG_ASFLAGS)
	endif
endif

ifeq ($(PLATFORM),Windows)
	libname = $(LIB_DIR)/libarchcode.lib
else
	libname = $(LIB_DIR)/libarchcode.a
endif

###############################################################################
# Commands
###############################################################################

.PHONY : clean .FORCE
.FORCE:

$(OBJ_DIR)/%.o: %.S
	$(CC) $(ASFLAGS) -c $< -o $@

$(OBJ_DIR)/%.o: %.c
	$(CC) $(CFLAGS) -c $< -o $@

$(libname): $(OBJ_DIR)/archcode.o
	cp $(OBJ_DIR)/archcode.o $(libname)

all: $(libname) go_bindings

go_bindings: $(libname)
	cp $(OBJ_DIR)/archcode.o $(BASE_DIR)/archcode/archcode.syso
	go build -o $(BIN_DIR)/arch_name $(BASE_DIR)

clean:
	-rm -f archcode.pc $(BASE_DIR)/archcode/archcode.syso
	-rm -rf $(OUT_DIR)

ifeq ($(PREFIX),)
PREFIX := /usr
endif

archcode.pc: .FORCE
	@echo 'prefix='$(PREFIX) > archcode.pc
	@echo 'exec_prefix=$${prefix}' >> archcode.pc
	@echo 'libdir=$${prefix}/lib' >> archcode.pc
	@echo 'includedir=$${prefix}/include' >> archcode.pc
	@echo '' >> archcode.pc
	@echo 'Name: archcode' >> archcode.pc
	@echo 'Description: Asssign an int32 to the architecture' >> archcode.pc
	@echo 'Version: '$(VERSION) >> archcode.pc
	@echo 'URL: https://github.com/prysmaticlabs/archcode' >> archcode.pc
	@echo 'LIBS: -L$${libdir} -larchcode' >> archcode.pc
	@echo 'Cflags: -I$${includedir}'>> archcode.pc

ifneq ($(PLATFORM),Windows)
install: $(libname) archcode.pc
	install -d $(DESTDIR)$(PREFIX)/lib
	install -m 644 $(libname) $(DESTDIR)$(PREFIX)/lib/
	install -d $(DESTDIR)$(PREFIX)/include
	install -m 644 archcode.h $(DESTDIR)$(PREFIX)/include/
	install -d $(DESTDIR)$(PREFIX)/lib/pkgconfig
	install -m 644 archcode.pc $(DESTDIR)$(PREFIX)/lib/pkgconfig/archcode.pc

uninstall: $(libname)
	rm $(DESTDIR)$(PREFIX)/lib/libarchcode.a
endif
