package tencentcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// go test -i; go test -test.run TestAccTencentCloudKmsPublicKeyDataSource_basic -v
func TestAccTencentCloudKmsPublicKeyDataSource_basic(t *testing.T) {
	t.Parallel()
	rName := fmt.Sprintf("tf-testacc-kms-key-%s", acctest.RandString(13))
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccKmsPublicKeyDataSource, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTencentCloudDataSourceID("data.tencentcloud_kms_public_key.example"),
				),
			},
		},
	})
}

const testAccKmsPublicKeyDataSource = `
data "tencentcloud_kms_public_key" "example" {
  key_id = tencentcloud_kms_key.example.id
}

resource "tencentcloud_kms_key" "example" {
  alias                         = "%s"
  description                   = "example of kms key"
  key_usage                     = "ASYMMETRIC_DECRYPT_RSA_2048"
  is_enabled                    = true
}
`
