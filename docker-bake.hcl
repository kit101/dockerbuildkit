variable "TAGS" {
}

group "default" {
  targets = [
    "test"
  ]
}

target "test" {
    context     = "."
    tags        = formatlist("kit101z/imagename:%s", compact(split(",", "${TAGS}")))
    platforms   = ["linux/amd64", "linux/arm64"]
    dockerfile  = "./docker/Dockerfile"
    labels      = {
        "com.cqcyit.container.build-time" = timestamp()
    }
}