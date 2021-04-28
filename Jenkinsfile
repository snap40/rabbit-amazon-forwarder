pipeline {
    agent {
        label 'linux && docker'
    }

    options {
        timestamps()
    }

    environment {
        IMAGE_NAME = "rmq-forwarder"
        IMAGE_VERSION = "${env.GIT_COMMIT}-${env.BUILD_ID}"
    }

    stages {
        stage("Build") {
            steps {
                script {
                    def image = docker.build("${env.IMAGE_NAME}:${env.IMAGE_VERSION}", ".")
                }
            }
        }

        stage("Test") {
            agent {
                docker {
                    image "golang:1.16.3-alpine3.13"
                    args '--tmpfs /.config -u root:root'
                }
            }
            steps {
                sh 'CGO_ENABLED=0 go test -json ./...'
            }
        }

        stage("Publish") {
            steps {
                script {
                    def image = docker.image("${env.IMAGE_NAME}:${env.IMAGE_VERSION}")
                    docker.withRegistry("https://${env.OPS_AWS_ACCOUNT_ID}.dkr.ecr.eu-west-1.amazonaws.com") {
                        image.push()
                    }
                    docker.withRegistry("https://${env.OPS_AWS_ACCOUNT_ID}.dkr.ecr.us-west-2.amazonaws.com") {
                        image.push()
                    }
                }
            }
        }

        stage("Publish Release Tag") {
            when { buildingTag() }
            steps {
                script {
                    def image = docker.image("${env.IMAGE_NAME}:${env.IMAGE_VERSION}")
                    docker.withRegistry("https://${env.OPS_AWS_ACCOUNT_ID}.dkr.ecr.eu-west-1.amazonaws.com") {
                        image.push(env.TAG_NAME)
                    }
                    docker.withRegistry("https://${env.OPS_AWS_ACCOUNT_ID}.dkr.ecr.us-west-2.amazonaws.com") {
                        image.push(env.TAG_NAME)
                    }
                }
            }
        }
    }
}