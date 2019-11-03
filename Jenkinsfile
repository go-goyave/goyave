pipeline {
    agent any
    stages {
        stage('Checkout'){
            steps {
                checkout scm
            }
        }

        stage('Test'){
            steps {
                sh 'go test -v -coverprofile=coverage.out -coverpkg=./... ./...'
                sh 'go tool cover -func=coverage.out | grep total'
                sh 'rm coverage.out'
            }
        }

        stage('Build Docs'){
            steps {
                dir('docs_src') {
                    sh 'npm install'
                    sh 'npm run build'
                    // Commit
                }
            }
        }

        stage('Cleanup'){
            steps {
                dir('docs_src') {
                    sh 'npm prune'
                    sh 'rm node_modules -rf'
                }
            }
        }
    }
}