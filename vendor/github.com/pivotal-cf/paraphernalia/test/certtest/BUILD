load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["certtest.go"],
    visibility = ["//visibility:public"],
    deps = ["@com_github_square_certstrap//pkix:go_default_library"],
)
