# kyma-operator-mothership

POC for the repository layout of the operator-based Kyma reconciliation project.

The project hosts a golang module for the mothership operator. The implementation is just a dummy, but proves that the project/repository split works as expected.

This operator reacts on k8s APIs events for the types defined within the project, but it's also capable of creating/updating k8s objects of types managed by other operators - imported from other projects.

For the purpose of POC this operator acts on dummy type `Kyma`. Upon receiving an event related to the observed `Kyma` type, it creates/deletes instances of `IstioConfiguration` type imported from https://github.com/Tomasz-Smelcerz-SAP/kyma-operator-istio repository.

Steps:

1) Setup the github repository

2) Create subdirectory for the operator

   `mkdir operator`

3) Generate the operator project

    `cd operator`
    
    `go mod init`
    
    `kubebuilder init --domain kyma-project.io`

    **Note: Generate both the Resource and the Controller**

    `kubebuilder create api --group inventory --version v1alpha1 --kind Kyma`

    `make manifests`

4) Push changes to github

5) Write the code that adds integration with IstioConfiguration objects

   *Note: Look at the commits that introduce operator implementation*

  - Creation of `Kyma` object should result in creation of `IstioConfiguration`
  - Deletion of `Kyma` object should result in deletion of related `IstioConfiguration`, if it exists.


