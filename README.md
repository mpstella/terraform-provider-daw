### Experimental DAW TF provider

#### Setup

edit your ~/.terraformrc to include below (where PATH = go binary path)

```
provider_installation {

  dev_overrides {
      "stella.com/anz/daw" = "<PATH>"
  }

  direct {}
}
```

##### Debug/Trace

```
$> export TF_LOG_PATH=terraform_trace.log
$> export TF_LOG=TRACE
```

Ensure you can login to GCP (easiest it to us adc)

```text
$> gcloud auth application-default login
```

##### Create new release
```
$> git tag -a v?.?.? -m "Release version v?.?.?"
$> git push origin v?.?.?
 
```


---

### Some concepts

[data sources](https://spacelift.io/blog/terraform-data-sources-how-they-are-utilised)
> Resources in Terraform represent the infrastructure components we want to create, manage, or delete. 
> They define the desired state of a particular resource type, such as virtual machines, databases, networks, and more. 
> Resources are the building blocks of our infrastructure and directly interact with our cloud provider’s APIs to create or modify actual resources.

> Data sources are used to query information from existing resources or external systems. 
> They provide a way to fetch specific attributes or data that we need to incorporate into our Terraform configuration. 
> Data sources do not create or manage resources; they retrieve information to inform the configuration of your resources.