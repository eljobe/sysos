# Feature Request: Support syso files as source files in rules_go

The purpose of this repository is to explain in detail why
http:/github.com/bazel-contrib/rules_go should support "syso" files in
the src lists for `go_library` and `go_test` rules in addition to the
current support in `go_binary` rules.

## Rationale: Feature parity with `go build`

Currently, go build links `syso` files into package libraries and
binaries (including test binaries) during a build.

As a quick proof that this is happenig with `go build` invocations and
`go test` invocations, I have run the following steps:

### 1. Perform a go test

This command causes the build to rebuild all the libraries and
binaries (`-a`) it needs instead of relying on the build cache, and it
prints out a log of all of the go tool commands it is running (`-x`)
as well as the output of those tools (`-v`). `-work` tells it to
retain the contents of the working directory and all of the
intermittent files that are created therein.

```
❯ go test -a -x -v --work ./... &> ~/tmp/sysos-go-test.log
```

This command just shows that the test actually passed:

```
❯ tail -3 ~/tmp/sysos-go-test.log
--- PASS: TestGetArchName (0.00s)
PASS
ok      github.com/eljobe/sysos/archcode        0.002s
```

### 2. Inspect the packages for archcode library and archcode_test binary.

These two commands show me where the working directory is, and the two
`ar` (archive creation calls) that go uses to make the library and
test binary. The main reason I need those two calls is to know which
directories I can inspect to find the symbols in the output packages.

```
❯ head -1 ~/tmp/sysos-go-test.log
WORK=/tmp/go-build3268603678
❯ grep "archcode_amd64.syso" ~/tmp/sysos-go-test.log
/home/linuxbrew/.linuxbrew/Cellar/go/1.23.2/libexec/pkg/tool/linux_amd64/pack r $WORK/b055/_pkg_.a $WORK/b055/wrap_arch_code_amd64.o ./archcode_amd64.syso # internal
/home/linuxbrew/.linuxbrew/Cellar/go/1.23.2/libexec/pkg/tool/linux_amd64/pack r $WORK/b057/_pkg_.a $WORK/b057/wrap_arch_code_amd64.o ./archcode_amd64.syso # internal
```

Once I know where the packages files can be found, I can list the
symbols in them with the `go tool nm` commands, and specifically
show that the symbolds from the `archcode_amd.syso` file have
been included in the packages.

```
❯ WORK=/tmp/go-build3268603678
❯ go tool nm ${WORK}/b055/_pkg_.a | grep archcode_amd64
/tmp/go-build3268603678/b055/_pkg_.a(archcode_amd64.s):        0 t
/tmp/go-build3268603678/b055/_pkg_.a(archcode_amd64.s):        0 _
/tmp/go-build3268603678/b055/_pkg_.a(archcode_amd64.s):        0 _
/tmp/go-build3268603678/b055/_pkg_.a(archcode_amd64.s):        0 _
/tmp/go-build3268603678/b055/_pkg_.a(archcode_amd64.s):        0 _
/tmp/go-build3268603678/b055/_pkg_.a(archcode_amd64.s):        0 _
/tmp/go-build3268603678/b055/_pkg_.a(archcode_amd64.s):        0 T GetArchCode
/tmp/go-build3268603678/b055/_pkg_.a(archcode_amd64.s):        0 _ archcode.c
❯ go tool nm ${WORK}/b057/_pkg_.a | grep archcode_amd64
/tmp/go-build3268603678/b057/_pkg_.a(archcode_amd64.s):        0 t
/tmp/go-build3268603678/b057/_pkg_.a(archcode_amd64.s):        0 _
/tmp/go-build3268603678/b057/_pkg_.a(archcode_amd64.s):        0 _
/tmp/go-build3268603678/b057/_pkg_.a(archcode_amd64.s):        0 _
/tmp/go-build3268603678/b057/_pkg_.a(archcode_amd64.s):        0 _
/tmp/go-build3268603678/b057/_pkg_.a(archcode_amd64.s):        0 _
/tmp/go-build3268603678/b057/_pkg_.a(archcode_amd64.s):        0 T GetArchCode
/tmp/go-build3268603678/b057/_pkg_.a(archcode_amd64.s):        0 _ archcode.c
```

### 3. Repeat the process for `go build` of the main program.

This just shows that go build also includes the smybols from the
`syso` files when building the library and the arch_name binary.

```
❯ go build -a -x -v --work -o arch_name ./main.go &> ~/tmp/sysos-go-build.log
❯ head -1 ~/tmp/sysos-go-build.log
WORK=/tmp/go-build2620151146
❯ WORK=/tmp/go-build2620151146
❯ grep "archcode_amd64.syso" ~/tmp/sysos-go-build.log
/home/linuxbrew/.linuxbrew/Cellar/go/1.23.2/libexec/pkg/tool/linux_amd64/pack r $WORK/b055/_pkg_.a $WORK/b055/wrap_arch_code_amd64.o ./archcode_amd64.syso # internal
❯ go tool nm ${WORK}/b055/_pkg_.a | grep archcode_amd64
/tmp/go-build2620151146/b055/_pkg_.a(archcode_amd64.s):        0 t
/tmp/go-build2620151146/b055/_pkg_.a(archcode_amd64.s):        0 _
/tmp/go-build2620151146/b055/_pkg_.a(archcode_amd64.s):        0 _
/tmp/go-build2620151146/b055/_pkg_.a(archcode_amd64.s):        0 _
/tmp/go-build2620151146/b055/_pkg_.a(archcode_amd64.s):        0 _
/tmp/go-build2620151146/b055/_pkg_.a(archcode_amd64.s):        0 _
/tmp/go-build2620151146/b055/_pkg_.a(archcode_amd64.s):        0 T GetArchCode
/tmp/go-build2620151146/b055/_pkg_.a(archcode_amd64.s):        0 _ archcode.c
❯ ./arch_name
Architecture: x86_64
❯ go tool nm arch_name | grep ArchCode
  401000 T GetArchCode
  48f220 T github.com/eljobe/sysos/archcode.GetArchCode.abi0
```

## Problem: Cannot reproduce with `bazel` and `rules_go`

There are three interesting `go_*` targets in this repository, and I
will describe two things about them:
   1. how each of them don't behave the same as the `go build` equivalants.
   2. any hacks that can produce correct (or nearly correcct)  behavior.

The three targest are:
   * `//archcode:go_default_library`
   * `//archcode:go_default_test`
   * `//arch_name`
  
### Analysis of `//archcode:go_default_library`

Here is the ideal way we'd like to be able to get the
`go_default_library` target to consume previously built (or built as
part of the `bazel build`) `.syso` files.

`archcode/BUILD.bazel`
```
go_library(
    name = "go_default_library",
    srcs = [
        "archcode.go",
        "wrap_arch_code_amd64.s",
        "archcode_amd64.syso", # source of error
    ],
    importpath = "github.com/eljobe/sysos/archcode",
    visibility = ["//visibility:public"],
)
```

#### How it breaks

`.syso` files are simply not allowed as sources despite their being
listed in the `go build` documentation among the file types which are
valid sorouces for a `go build` invocation at least since go1.4
(I couldn't find older docs.)

https://pkg.go.dev/cmd/go@go1.4#hdr-File_types


```
❯ bazel build //archcode:go_default_library
Computing main repo mapping:
Loading:
Loading: 0 packages loaded
bazel: Entering directory `/home/pepper/.cache/bazel/_bazel_pepper/2e19539aee4643ad982cb467683c8015/execroot/_main/'
Analyzing: target //archcode:go_default_library (1 packages loaded, 0 targets configured)
Analyzing: target //archcode:go_default_library (1 packages loaded, 0 targets configured)
[0 / 1] [Prepa] BazelWorkspaceStatusAction stable-status.txt
ERROR: /home/pepper/dev/github.com/eljobe/sysos/archcode/BUILD.bazel:3:11: in srcs attribute of go_library rule //archcode:go_default_library: source file '//archcode:archcode_amd64.syso' is misplaced here (expected .go, .s, .S, .h, .c, .cc, .cpp, .cxx, .h, .hh, .hpp, .hxx, .inc, .m or .mm). Since this rule was created by the macro 'go_library_macro', the error might have been caused by the macro implementation
ERROR: /home/pepper/dev/github.com/eljobe/sysos/archcode/BUILD.bazel:3:11: in srcs attribute of go_library rule //archcode:go_default_library: '//archcode:archcode_amd64.syso' does not produce any go_library srcs files (expected .go, .s, .S, .h, .c, .cc, .cpp, .cxx, .h, .hh, .hpp, .hxx, .inc, .m or .mm). Since this rule was created by the macro 'go_library_macro', the error might have been caused by the macro i
mplementation
ERROR: /home/pepper/dev/github.com/eljobe/sysos/archcode/BUILD.bazel:3:11: Analysis of target '//archcode:go_default_library' failed
bazel: Leaving directory `/home/pepper/.cache/bazel/_bazel_pepper/2e19539aee4643ad982cb467683c8015/execroot/_main/'
ERROR: Analysis of target '//archcode:go_default_library' failed; build aborted
INFO: Elapsed time: 0.152s, Critical Path: 0.00s
INFO: 1 process: 1 internal.
ERROR: Build did NOT complete successfully
```

#### How can **almost** be hacked to work

If you remove the `.syso` source file and pretend that it's using
`cgo` and take a dependency on the `//src:archcode` library in
`cdeps`, the go library will actually "build" (instead of erroring
out.)

`archcode/BUILD.bazel`
```
go_library(
    name = "go_default_library",
    srcs = [
        "archcode.go",
        "wrap_arch_code_amd64.s",
        # "archcode_amd64.syso", # source of error
    ],
    # Hack: No symbols :(
    cgo = True,
    cdeps = ["//src:archcode"],
    importpath = "github.com/eljobe/sysos/archcode",
    visibility = ["//visibility:public"],
)
```

Now, the `go_default_library` target will build.

```
❯ bazel build //archcode:go_default_library
Computing main repo mapping:
Loading:
Loading: 0 packages loaded
bazel: Entering directory `/home/pepper/.cache/bazel/_bazel_pepper/2e19539aee4643ad982cb467683c8015/execroot/_main/'
Analyzing: target //archcode:go_default_library (1 packages loaded, 0 targets configured)
Analyzing: target //archcode:go_default_library (1 packages loaded, 0 targets configured)
[0 / 1] [Prepa] BazelWorkspaceStatusAction stable-status.txt
INFO: Analyzed target //archcode:go_default_library (1 packages loaded, 3 targets configured).
INFO: Found 1 target...
bazel: Leaving directory `/home/pepper/.cache/bazel/_bazel_pepper/2e19539aee4643ad982cb467683c8015/execroot/_main/'
Target //archcode:go_default_library up-to-date:
  bazel-bin/archcode/go_default_library.x
  INFO: Elapsed time: 0.178s, Critical Path: 0.00s
  INFO: 1 process: 1 internal.
  INFO: Build completed successfully, 1 total action
```

**But** when you list the symbols in the resulting go object file, the
GetArchCode method from archcode.c is nowhere to be found.

```
❯ go tool nm  bazel-out/k8-fastbuild/bin/archcode/go_default_library.a
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(_go_.o):                U
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(_go_.o):                U
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(_go_.o):                U
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(_go_.o):                U
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(_go_.o):                U
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(_go_.o):                U <autogenerated>
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(_go_.o):                U archcode/archcode.go
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(_go_.o):            fc8 R gclocals·g2BeySu+wFnoycgXfElmcg==
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(_go_.o):                U github.com/eljobe/sysos/archcode.GetArchCode
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(_go_.o):            fd0 T github.com/eljobe/sysos/archcode.GetArchCode
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(_go_.o):            daf R github.com/eljobe/sysos/archcode.GetArchCode.arginfo0
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(_go_.o):            da5 R github.com/eljobe/sysos/archcode.GetArchCode.args_stackmap
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(_go_.o):            c07 T github.com/eljobe/sysos/archcode.GetArchName
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(_go_.o):            cb1 ? go:constinfo.github.com/eljobe/sysos/archcode
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(_go_.o):            ff9 ? go:cuinfo.packagename.github.com/eljobe/sysos/archcode
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(_go_.o):            ff3 ? go:cuinfo.producer.github.com/eljobe/sysos/archcode
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(_go_.o):                U go:info.int
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(_go_.o):                U go:info.int32
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(_go_.o):                U go:info.string
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(_go_.o):            f30 R go:string."ARM"
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(_go_.o):            f33 R go:string."ARM64"
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(_go_.o):            fb4 R go:string."Unknown Architecture"
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(_go_.o):            f2d R go:string."x86"
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(_go_.o):            f27 R go:string."x86_64"
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(s0.o):          U
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(s0.o):          U
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(s0.o):          U GetArchCode
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(s0.o):          U archcode/wrap_arch_code_amd64.s
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(s0.o):      467 T github.com/eljobe/sysos/archcode.GetArchCode
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(s0.o):          U github.com/eljobe/sysos/archcode.GetArchCode.arginfo0
bazel-out/k8-fastbuild/bin/archcode/go_default_library.a(s0.o):          U github.com/eljobe/sysos/archcode.GetArchCode.args_stackmap
```

### Analysis of `//archcode:go_default_test`

Here is the ideal way we'd like to be able to get the
`go_default_test` target to consume previously built (or built as
part of the `bazel build`) `.syso` files.

`archcode/BUILD.bazel`
```
go_test(
    name = "go_default_test",
    srcs = ["archcode_test.go"],
    embed = [":go_default_library"],
)
```

Now, it might not be obvious right away, but, `embed` is essentially
saying, "Include the `srcs` from these targets in this target's
`srcs`." So, transitively, this would consume our ideal
`archcode_amd64.syso` from the `:go_default_library` target in the
same package.

#### How it breaks

Again, the `.syso` files aren't allowed in the `srcs` of a
`go_library` rule, so, we cannot even get to the error in the test
target.

```
❯ bazel build //archcode:go_default_test
Computing main repo mapping:
Loading:
Loading: 0 packages loaded
bazel: Entering directory `/home/pepper/.cache/bazel/_bazel_pepper/2e19539aee4643ad982cb467683c8015/execroot/_main/'
Analyzing: target //archcode:go_default_test (1 packages loaded, 0 targets configured)
Analyzing: target //archcode:go_default_test (1 packages loaded, 0 targets configured)
[0 / 1] [Prepa] BazelWorkspaceStatusAction stable-status.txt
ERROR: /home/pepper/dev/github.com/eljobe/sysos/archcode/BUILD.bazel:3:11: in srcs attribute of go_library rule //archcode:go_default_library: source file '//archcode:archcode_amd64.syso' is misplaced here (expected .go, .s, .S, .h, .c, .cc, .cpp, .cxx, .h, .hh, .hpp, .hxx, .inc, .m or .mm). Since this rule was created by the macro 'go_library_macro', the error might have been caused by the macro implementation
ERROR: /home/pepper/dev/github.com/eljobe/sysos/archcode/BUILD.bazel:3:11: in srcs attribute of go_library rule //archcode:go_default_library: '//archcode:archcode_amd64.syso' does not produce any go_library srcs files (expected .go, .s, .S, .h, .c, .cc, .cpp, .cxx, .h, .hh, .hpp, .hxx, .inc, .m or .mm). Since this rule was created by the macro 'go_library_macro', the error might have been caused by the macro i
mplementation
ERROR: /home/pepper/dev/github.com/eljobe/sysos/archcode/BUILD.bazel:3:11: Analysis of target '//archcode:go_default_library' failed
bazel: Leaving directory `/home/pepper/.cache/bazel/_bazel_pepper/2e19539aee4643ad982cb467683c8015/execroot/_main/'
ERROR: Analysis of target '//archcode:go_default_test' failed; build aborted: Analysis failed
INFO: Elapsed time: 0.148s, Critical Path: 0.00s
INFO: 1 process: 1 internal.
ERROR: Build did NOT complete successfully
```

#### How it **cannot** even be hacked to work.

However, if we switch the `go_default_library` definition to the one
with `cgo = True` then we get a different error when building the
`//archcode:go_default_test` target.

```
❯ bazel build //archcode:go_default_test
Computing main repo mapping:
Loading:
Loading: 0 packages loaded
bazel: Entering directory `/home/pepper/.cache/bazel/_bazel_pepper/2e19539aee4643ad982cb467683c8015/execroot/_main/'
Analyzing: target //archcode:go_default_test (1 packages loaded, 0 targets configured)
Analyzing: target //archcode:go_default_test (1 packages loaded, 0 targets configured)
[0 / 1] [Prepa] BazelWorkspaceStatusAction stable-status.txt
INFO: Analyzed target //archcode:go_default_test (1 packages loaded, 5 targets configured).
ERROR: /home/pepper/dev/github.com/eljobe/sysos/archcode/BUILD.bazel:17:8: GoLink archcode/go_default_test_/go_default_test failed: (Exit 1): builder failed: error executing GoLink command (from target //archcode:go_default_test) bazel-out/k8-opt-exec-ST-13d3ddad9198/bin/external/rules_go~~go_sdk~sysos__download_0/builder_reset/builder link -sdk external/rules_go~~go_sdk~sysos__download_0 -installsuffix linux_a
md64 -arc ... (remaining 29 arguments skipped)

Use --sandbox_debug to see verbose messages from the sandbox and retain the sandbox build root for debugging
github.com/eljobe/sysos/archcode.GetArchCode: relocation target GetArchCode not defined
link: error running subcommand external/rules_go~~go_sdk~sysos__download_0/pkg/tool/linux_amd64/link: exit status 2
bazel: Leaving directory `/home/pepper/.cache/bazel/_bazel_pepper/2e19539aee4643ad982cb467683c8015/execroot/_main/'
Target //archcode:go_default_test failed to build
Use --verbose_failures to see the command lines of failed build steps.
INFO: Elapsed time: 0.283s, Critical Path: 0.13s
INFO: 2 processes: 2 internal.
ERROR: Build did NOT complete successfully
```

This tells us that the `relocation target GetArchCode` is not
declared. I belive this is because of the same reason we had to add
the `cgo` and `cdeps` to the `go_default_library` file. Essentially,
the c function's symbol isn't available to link to.

One idea is to try and add the same `cgo = True` and `cdeps =
[//src:archcode]` that "fixed" the `go_default_library` rule.

But, there is code in the implementation of `embed` which makes sure
that there can only be a single `cgo = True` target in any chain of
embeds.

If instead, we try to add all of the sources explicitly needed by the
`go_default_test` rule to the target's `src` parameter, and add the
`cgo` and `cdeps` arguments in an attempt to get the test binary to
link in the mising symbol, it still doesn't work (I'm not totally
clear on why.)

`archcode/BUILD.bazel`
```
go_test(
    name = "go_default_test",
    srcs = [
        "archcode_test.go",
        "archcode.go",
        "wrap_arch_code_amd64.s",
    ],
    cgo = True,
    cdeps = ["//src:archcode"],
    # embed = [":go_default_library"],
)
```

```
❯ bazel build //archcode:go_default_test

Computing main repo mapping:
Loading:
Loading: 0 packages loaded
bazel: Entering directory `/home/pepper/.cache/bazel/_bazel_pepper/2e19539aee4643ad982cb467683c8015/execroot/_main/'
Analyzing: target //archcode:go_default_test (0 packages loaded, 0 targets configured)
Analyzing: target //archcode:go_default_test (0 packages loaded, 0 targets configured)
[0 / 1] [Prepa] BazelWorkspaceStatusAction stable-status.txt
INFO: Analyzed target //archcode:go_default_test (0 packages loaded, 0 targets configured).
ERROR: /home/pepper/dev/github.com/eljobe/sysos/archcode/BUILD.bazel:17:8: GoLink archcode/go_default_test_/go_default_test failed: (Exit 1): builder failed: error executing GoLink command (from target //archcode:go_default_test) bazel-out/k8-opt-exec-ST-13d3ddad9198/bin/external/rules_go~~go_sdk~sysos__download_0/builder_reset/builder link -sdk external/rules_go~~go_sdk~sysos__download_0 -installsuffix linux_a
md64 -arc ... (remaining 29 arguments skipped)

Use --sandbox_debug to see verbose messages from the sandbox and retain the sandbox build root for debugging
github.com/eljobe/sysos/archcode.GetArchCode: relocation target GetArchCode not defined
link: error running subcommand external/rules_go~~go_sdk~sysos__download_0/pkg/tool/linux_amd64/link: exit status 2
bazel: Leaving directory `/home/pepper/.cache/bazel/_bazel_pepper/2e19539aee4643ad982cb467683c8015/execroot/_main/'
Target //archcode:go_default_test failed to build
Use --verbose_failures to see the command lines of failed build steps.
INFO: Elapsed time: 0.210s, Critical Path: 0.10s
INFO: 2 processes: 2 internal.
ERROR: Build did NOT complete successfully
```

### Analysis of `//:arch_name`

Here is the ideal way we'd like to be able to get the
`arch_name` target to consume previously built (or built as
part of the `bazel build`) `.syso` files.

`BUILD.bazel`
```
go_binary(
    name = "arch_name",
    srcs = [
        "main.go",
     ],
    deps = ["//archcode:go_default_library"],
)
```

That is to say, for this repository, we want the `archcode` package to
just create a go object archive that already has all the symbols that
are needed by the `arch_name` binary.

In other repositories, however, all of the source code for a given
binary might reside in the same package. And, in that case, it would
be good to support `.syso` directly in the `srcs` of the `go_binary`
rule.  And, in fact, it IS supported there. It was added this PR
https://github.com/bazel-contrib/rules_go/pull/3763 so that Windows
programs could include the version and icon info into Widows
executables.

So, while `go_binary` rules have support for `.syso` files in `srcs`
we also want to ensure, while implementing this feature request that
`go_binary` targets can also get the symbols they need through
transitive `go_library` dependencies on targets which were packaged
with the `.syso` files already.

#### How it breaks

Again, the `.syso` files aren't allowed in the `srcs` of a
`go_library` rule, so, we cannot even get to the error in the test
target.

```
❯ bazel build //:arch_name
Computing main repo mapping:
Loading:
Loading: 0 packages loaded
bazel: Entering directory `/home/pepper/.cache/bazel/_bazel_pepper/2e19539aee4643ad982cb467683c8015/execroot/_main/'
Analyzing: target //:arch_name (0 packages loaded, 0 targets configured)
Analyzing: target //:arch_name (0 packages loaded, 0 targets configured)
[0 / 1] [Prepa] BazelWorkspaceStatusAction stable-status.txt
ERROR: /home/pepper/dev/github.com/eljobe/sysos/archcode/BUILD.bazel:3:11: in srcs attribute of go_library rule //archcode:go_default_library: source file '//archcode:archcode_amd64.syso' is misplaced here (expected .go, .s, .S, .h, .c, .cc, .cpp, .cxx, .h, .hh, .hpp, .hxx, .inc, .m or .mm). Since this rule was created by the macro 'go_library_macro', the error might have been caused by the macro implementation
ERROR: /home/pepper/dev/github.com/eljobe/sysos/archcode/BUILD.bazel:3:11: in srcs attribute of go_library rule //archcode:go_default_library: '//archcode:archcode_amd64.syso' does not produce any go_library srcs files (expected .go, .s, .S, .h, .c, .cc, .cpp, .cxx, .h, .hh, .hpp, .hxx, .inc, .m or .mm). Since this rule was created by the macro 'go_library_macro', the error might have been caused by the macro i
mplementation
ERROR: /home/pepper/dev/github.com/eljobe/sysos/archcode/BUILD.bazel:3:11: Analysis of target '//archcode:go_default_library' failed
bazel: Leaving directory `/home/pepper/.cache/bazel/_bazel_pepper/2e19539aee4643ad982cb467683c8015/execroot/_main/'
ERROR: Analysis of target '//:arch_name' failed; build aborted: Analysis failed
INFO: Elapsed time: 0.109s, Critical Path: 0.00s
INFO: 1 process: 1 internal.
ERROR: Build did NOT complete successfully
```

If we remove the `.syso` file from the `//archcode:go_default_library`
`srcs` and build with the `cgo` options, the build still fails the
same with it did for the `//archcode:go_default_test` target.

```
❯ bazel build //:arch_name
Computing main repo mapping:
Loading:
Loading: 0 packages loaded
bazel: Entering directory `/home/pepper/.cache/bazel/_bazel_pepper/2e19539aee4643ad982cb467683c8015/execroot/_main/'
Analyzing: target //:arch_name (0 packages loaded, 0 targets configured)
Analyzing: target //:arch_name (0 packages loaded, 0 targets configured)
[0 / 1] [Prepa] BazelWorkspaceStatusAction stable-status.txt
INFO: Analyzed target //:arch_name (1 packages loaded, 4 targets configured).
ERROR: /home/pepper/dev/github.com/eljobe/sysos/BUILD.bazel:14:10: GoLink arch_name_/arch_name failed: (Exit 1): builder failed: error executing GoLink command (from target //:arch_name) bazel-out/k8-opt-exec-ST-13d3ddad9198/bin/external/rules_go~~go_sdk~sysos__download_0/builder_reset/builder link -sdk external/rules_go~~go_sdk~sysos__download_0 -installsuffix linux_amd64 -arc ... (remaining 19 arguments skipp
ed)

Use --sandbox_debug to see verbose messages from the sandbox and retain the sandbox build root for debugging
github.com/eljobe/sysos/archcode.GetArchCode: relocation target GetArchCode not defined
link: error running subcommand external/rules_go~~go_sdk~sysos__download_0/pkg/tool/linux_amd64/link: exit status 2
bazel: Leaving directory `/home/pepper/.cache/bazel/_bazel_pepper/2e19539aee4643ad982cb467683c8015/execroot/_main/'
Target //:arch_name failed to build
Use --verbose_failures to see the command lines of failed build steps.
INFO: Elapsed time: 0.205s, Critical Path: 0.06s
INFO: 2 processes: 2 internal.
ERROR: Build did NOT complete successfully
```

#### How it can be hacked to work

Now, because the `go_binary` rule actually does allow `.syso` files
among its `srcs`, with the `cgo` hack in place to get the
`//archcode:go_default_library` to build (even though it dosn't have
the symbol we need in it,) we can also incude the
`archcode_arm64.syso` from the `prebuilt` directory to the sources and
it will provide the missing symbol.


`BUILD.bazel`
```
go_binary(
    name = "arch_name",
    srcs = [
        "main.go",
        "prebuilt/archcode_amd64.syso",
    ],
    deps = ["//archcode:go_default_library"],
)
```

I still consider this a "hack" because in the ideal setup for the
repository, the `//archcode:go_default_library` would have provided
the symbol transitively.

But, at least it gets he binary to build and run correctly.

```
❯ bazel run //:arch_name
Computing main repo mapping:
Loading:
Loading: 0 packages loaded
bazel: Entering directory `/home/pepper/.cache/bazel/_bazel_pepper/2e19539aee4643ad982cb467683c8015/execroot/_main/'
Analyzing: target //:arch_name (0 packages loaded, 0 targets configured)
Analyzing: target //:arch_name (0 packages loaded, 0 targets configured)
[0 / 1] [Prepa] BazelWorkspaceStatusAction stable-status.txt
INFO: Analyzed target //:arch_name (0 packages loaded, 0 targets configured).
INFO: Found 1 target...
bazel: Leaving directory `/home/pepper/.cache/bazel/_bazel_pepper/2e19539aee4643ad982cb467683c8015/execroot/_main/'
Target //:arch_name up-to-date:
  bazel-bin/arch_name_/arch_name
  INFO: Elapsed time: 0.099s, Critical Path: 0.00s
  INFO: 1 process: 1 internal.
  INFO: Build completed successfully, 1 total action
  INFO: Running command line: bazel-bin/arch_name_/arch_name
  Architecture: x86_64
```

It is also worth noting that the binary package (before it was
stripped of symbols) did actually contain the symbols from the
`prebuilt/archnode.syso` file.

```
❯ go tool nm  bazel-out/k8-fastbuild/bin/arch_name.a | grep archcode_amd64
bazel-out/k8-fastbuild/bin/arch_name.a(archcode_amd64.s):              0 t
bazel-out/k8-fastbuild/bin/arch_name.a(archcode_amd64.s):              0 _
bazel-out/k8-fastbuild/bin/arch_name.a(archcode_amd64.s):              0 _
bazel-out/k8-fastbuild/bin/arch_name.a(archcode_amd64.s):              0 _
bazel-out/k8-fastbuild/bin/arch_name.a(archcode_amd64.s):              0 _
bazel-out/k8-fastbuild/bin/arch_name.a(archcode_amd64.s):              0 _
bazel-out/k8-fastbuild/bin/arch_name.a(archcode_amd64.s):              0 T GetArchCode
bazel-out/k8-fastbuild/bin/arch_name.a(archcode_amd64.s):              0 _ archcode.c
```

This indicates that, if we use the existing `GoCompilePkg` action that
is already working for the `go_binary` rule that we should end up with
go object archives with all the desired symbols in them.

## Solution: Design ideas

There are actually two objectives, and they are ordered (both in terms
of importance and dependency.)
   1. Support `.syso` files: Allow them to appear in the `srcs`
      attribute of `go_test` and `go_library` rules.
   2. Produce `.syso` files: Allow users to generate `.syso` files
      from `cc_library` rules.
   
### Support `.syso` files

This should be trivial given the groundwork already laid in
https://github.com/bazel-contrib/rules_go/pull/3763.

Simply allow the `.syso` files in the `srcs` for `go_library`,
`go_test` and `go_source` and you're done. The code in `archive.bzl`
and `compliepkg.go` is already doing the "heavy lifting" correctly.

### Produce `.syso` files

I'm a little less sure how to do this correctly. I suppose there are a
couple of options for how this could look:

#### Option 1: New Rule

`BUILD.bazel`
```
cc_library( name = "clib", ...)

go_system_object_lib(
  name = "sysolib",
  deps = [":clib"],
)
# This would have outs = ["sysolib.syso"]

go_library(
    name = "go_default_library",
	srcs = [
		"some.go",
		":sysolib",
	],
)
```

#### Option 2: New attribute

`BUILD.bazel`
```
cc_library( name = "clib", ...)

go_library(
    name = "go_default_library",
	srcs = [
		"some.go",
	],
	syso_src_cdeps = [
	    ":clib",
	],
)
```

#### Analysis

I think I prefer the usability of the first one better. For some
reason, I think it's easier to reason about. But, I'm aware that
that's pretty "subjective."

