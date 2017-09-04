load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "info.go",
        "runner.go",
        "uptime.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "@com_github_tedsuo_ifrit//:go_default_library",
        "@com_github_tedsuo_ifrit//http_server:go_default_library",
        "@org_golang_x_net//trace:go_default_library",
    ],
)

go_test(
    name = "go_default_xtest",
    srcs = [
        "admin_suite_test.go",
        "runner_test.go",
    ],
    deps = [
        ":go_default_library",
        "@com_github_onsi_ginkgo//:go_default_library",
        "@com_github_onsi_gomega//:go_default_library",
        "@com_github_tedsuo_ifrit//:go_default_library",
    ],
    size = "small",
)
