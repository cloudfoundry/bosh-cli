Vagrant.configure('2') do |config|
  config.vm.provider :virtualbox do |v, override|
    override.vm.box = 'bosh-lite-ubuntu-trusty-virtualbox-293'
    override.vm.box_url = 'http://d3a4sadvqj176z.cloudfront.net/bosh-lite-virtualbox-ubuntu-trusty-293.box'
  end

  [:vmware_fusion, :vmware_desktop, :vmware_workstation].each do |provider|
    config.vm.provider provider do |v, override|
      override.vm.box = 'bosh-lite-ubuntu-trusty-vmware-15'
      override.vm.box_url = 'https://d3a4sadvqj176z.cloudfront.net/bosh-lite-vmware-ubuntu-trusty-15.box'
    end
  end

  config.vm.provider :aws do |v, override|
    override.vm.box = 'bosh-lite-ubuntu-trusty-aws-174'
    override.vm.box_url = 'https://d3a4sadvqj176z.cloudfront.net/bosh-lite-aws-ubuntu-trusty-174.box'
  end

  config.vm.synced_folder File.expand_path('../tmp', __FILE__), '/vagrant'
end
