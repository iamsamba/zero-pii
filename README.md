# zero-pii

A quick and easy way to handle PII (Personally Identifiable Information).

## Table of Contents
- [Introduction](#introduction)
- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Contributing](#contributing)
- [License](#license)

## Introduction
`zero-pii` designed to help developers handle and manage PII in their applications. It provides API first approach to identify, mask, and securely store PII, ensuring compliance with privacy regulations.

## Features
- Identify PII in datasets
- Mask or anonymize PII
- Secure storage of PII
- Easy integration with existing applications

## Installation
## AWS EKS Deployment

### Build and Push Docker Image

```sh
aws ecr get-login-password --region <region> | docker login --username AWS --password-stdin <your-aws-account-id>.dkr.ecr.<region>.amazonaws.com
docker build -t <repo-name>:latest .
docker tag <repo-name>:latest <your-aws-account-id>.dkr.ecr.<region>.amazonaws.com/<repo-name>:latest
docker push <your-aws-account-id>.dkr.ecr.<region>.amazonaws.com/<repo-name>:latest
```

### Deploy to EKS

1. Update the image in `deployment.yaml` with your ECR image URI.
2. Apply the deployment:

```sh
kubectl apply -f deployment.yaml
```

3. The service will be exposed via a LoadBalancer. Get the external URL with:

```sh
kubectl get svc zero-pii-service
```
## GCP GKE Deployment

### Prerequisites

- [Google Cloud SDK](https://cloud.google.com/sdk/docs/install)
- A Google Cloud project with billing enabled
- [GKE API enabled](https://console.cloud.google.com/marketplace/product/google/container.googleapis.com)
- Docker installed and authenticated with Google Cloud

### Build and Push Docker Image to Google Container Registry (GCR)

```sh
# Set your project ID
export PROJECT_ID=<your-gcp-project-id>
export IMAGE_NAME=zero-pii

# Authenticate Docker to GCR
gcloud auth configure-docker

# Build the Docker image
docker build -t gcr.io/$PROJECT_ID/$IMAGE_NAME:latest .

# Push the image to GCR
docker push gcr.io/$PROJECT_ID/$IMAGE_NAME:latest
```

### Deploy to GKE

1. Create a GKE cluster (if you don't have one):

    ```sh
    gcloud container clusters create zero-pii-cluster --zone <your-zone>
    gcloud container clusters get-credentials zero-pii-cluster --zone <your-zone>
    ```

2. Update the `image` field in your `deployment.yaml` to use your GCR image URI:  
   `gcr.io/<your-gcp-project-id>/zero-pii:latest`

3. Apply the deployment:

    ```sh
    kubectl apply -f deployment.yaml
    ```

4. Expose your service (if not already exposed):

    ```sh
    kubectl expose deployment zero-pii --type=LoadBalancer --port 80 --target-port 8080
    ```

5. Get the external IP:

    ```sh
    kubectl get service zero-pii
    ```

---

**Be sure to update your `deployment.yaml` image field as described above.**

Let me know if you want the `deployment.yaml` example for GKE or further customization!## GCP GKE Deployment

### Prerequisites

- [Google Cloud SDK](https://cloud.google.com/sdk/docs/install)
- A Google Cloud project with billing enabled
- [GKE API enabled](https://console.cloud.google.com/marketplace/product/google/container.googleapis.com)
- Docker installed and authenticated with Google Cloud

### Build and Push Docker Image to Google Container Registry (GCR)

```sh
# Set your project ID
export PROJECT_ID=<your-gcp-project-id>
export IMAGE_NAME=zero-pii

# Authenticate Docker to GCR
gcloud auth configure-docker

# Build the Docker image
docker build -t gcr.io/$PROJECT_ID/$IMAGE_NAME:latest .

# Push the image to GCR
docker push gcr.io/$PROJECT_ID/$IMAGE_NAME:latest
```

### Deploy to GKE

1. Create a GKE cluster (if you don't have one):

    ```sh
    gcloud container clusters create zero-pii-cluster --zone <your-zone>
    gcloud container clusters get-credentials zero-pii-cluster --zone <your-zone>
    ```

2. Update the `image` field in your `deployment.yaml` to use your GCR image URI:  
   `gcr.io/<your-gcp-project-id>/zero-pii:latest`

3. Apply the deployment:

    ```sh
    kubectl apply -f deployment.yaml
    ```

4. Expose your service (if not already exposed):

    ```sh
    kubectl expose deployment zero-pii --type=LoadBalancer --port 80 --target-port 8080
    ```

5. Get the external IP:

    ```sh
    kubectl get service zero-pii
    ```

---

**Be sure to update your `deployment.yaml` image field as described above.**

Let me know if you want the `deployment.yaml` example for GKE or further customization!

## Contributing
Contributions are welcome! Please follow these steps to contribute:

1. Fork the repository.
2. Create a new branch (`git checkout -b feature-branch`).
3. Make your changes.
4. Commit your changes (`git commit -am 'Add new feature'`).
5. Push to the branch (`git push origin feature-branch`).
6. Create a new Pull Request.

Please make sure your code adheres to the project's coding standards and includes appropriate tests.

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.




