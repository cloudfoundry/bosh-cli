load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["id.go"],
    visibility = ["//visibility:public"],
    deps = ["@com_github_oklog_ulid//:go_default_library"],
)

go_test(
    name = "go_default_xtest",
    srcs = [
        "id_test.go",
        "requestid_suite_test.go",
    ],
    deps = [
        ":go_default_library",
        "@com_github_onsi_ginkgo//:go_default_library",
        "@com_github_onsi_gomega//:go_default_library",
    ],
    size = "small",
)
