package tencentcloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// go test -i; go test -test.run TestAccTencentCloudTdmqEnvironmentAttributesDataSource_basic -v
func TestAccTencentCloudTdmqEnvironmentAttributesDataSource_basic(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTdmqEnvironmentAttributesDataSource,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTencentCloudDataSourceID("data.tencentcloud_tdmq_environment_attributes.example"),
				),
			},
		},
	})
}

const testAccTdmqEnvironmentAttributesDataSource = `
data "tencentcloud_tdmq_environment_attributes" "example" {
  environment_id = tencentcloud_tdmq_namespace.example.environ_name
  cluster_id     = tencentcloud_tdmq_instance.example.id
}

resource "tencentcloud_tdmq_instance" "example" {
  cluster_name = "tf_example"
  remark       = "remark."
  tags         = {
    "createdBy" = "terraform"
  }
}

resource "tencentcloud_tdmq_namespace" "example" {
  environ_name = "tf_example"
  msg_ttl      = 300
  cluster_id   = tencentcloud_tdmq_instance.example.id
  remark       = "remark."
}
`
