load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["acl.go"],
    visibility = ["//visibility:public"],
    deps = ["@org_golang_google_grpc//credentials:go_default_library"],
)

go_test(
    name = "go_default_xtest",
    srcs = [
        "acl_test.go",
        "suite_test.go",
    ],
    deps = [
        ":go_default_library",
        "//serve/grpcrunner:go_default_library",
        "//test/certtest:go_default_library",
        "@com_github_onsi_ginkgo//:go_default_library",
        "@com_github_onsi_gomega//:go_default_library",
        "@com_github_tedsuo_ifrit//:go_default_library",
        "@com_github_tedsuo_ifrit//ginkgomon:go_default_library",
        "@org_cloudfoundry_code_lager//:go_default_library",
        "@org_cloudfoundry_code_lager//lagertest:go_default_library",
        "@org_golang_google_grpc//:go_default_library",
        "@org_golang_google_grpc//credentials:go_default_library",
        "@org_golang_google_grpc//grpclog:go_default_library",
        "@org_golang_google_grpc//test/grpc_testing:go_default_library",
        "@org_golang_x_net//context:go_default_library",
    ],
    size = "small",
    tags = [
        "noci", # TODO(cb): work out why this test suite does not run in travis
    ],
)
