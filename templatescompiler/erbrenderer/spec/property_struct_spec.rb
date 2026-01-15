require "spec_helper"
require "erb_renderer"

RSpec.describe "PropertyStruct" do
  describe "initialization and attribute access" do
    it "provides dynamic attribute access for hash keys" do
      ps = PropertyStruct.new(name: "test", value: 42)
      expect(ps.name).to eq("test")
      expect(ps.value).to eq(42)
    end

    it "converts string keys to symbols" do
      ps = PropertyStruct.new("name" => "test", "value" => 42)
      expect(ps.name).to eq("test")
      expect(ps.value).to eq(42)
    end

    it "supports nested attribute access" do
      ps = PropertyStruct.new(config: {database: {host: "localhost", port: 5432}})
      nested = ps.config
      expect(nested).to be_a(PropertyStruct)
      expect(nested.database).to be_a(PropertyStruct)
      expect(nested.database.host).to eq("localhost")
      expect(nested.database.port).to eq(5432)
    end

    it "handles arrays of hashes" do
      ps = PropertyStruct.new(servers: [{name: "web1", ip: "10.0.0.1"}, {name: "web2", ip: "10.0.0.2"}])
      servers = ps.servers
      expect(servers).to be_an(Array)
      expect(servers.length).to eq(2)
      expect(servers.first.name).to eq("web1")
      expect(servers.last.ip).to eq("10.0.0.2")
    end

    it "responds to method queries correctly" do
      ps = PropertyStruct.new(existing_key: "value")
      expect(ps.respond_to?(:existing_key)).to be true
      expect(ps.respond_to?(:nonexistent_key)).to be false
    end
  end

  describe "Ruby standard library method pass-through" do
    it "supports array operations like map" do
      ps = PropertyStruct.new(ports: [8080, 8081, 8082])
      expect(ps.ports.map(&:to_s)).to eq(["8080", "8081", "8082"])
    end

    it "supports string operations" do
      ps = PropertyStruct.new(url: "https://example.com")
      expect(ps.url.start_with?("https")).to be true
      expect(ps.url.split("://")).to eq(["https", "example.com"])
    end

    it "supports hash operations" do
      ps = PropertyStruct.new(config: {a: 1, b: 2, c: 3})
      expect(ps.config.keys.sort).to eq([:a, :b, :c])
      expect(ps.config.values.sum).to eq(6)
    end

    it "supports nil and empty checks" do
      ps = PropertyStruct.new(empty_string: "", nil_value: nil, filled: "data")
      expect(ps.empty_string.empty?).to be true
      expect(ps.nil_value.nil?).to be true
      expect(ps.filled.nil?).to be false
    end
  end

  describe "compatibility across Ruby versions" do
    it "works with ERB rendering" do
      template = ERB.new("<%= obj.name.upcase %>: <%= obj.ports.join(',') %>")
      ps = PropertyStruct.new(name: "service", ports: [80, 443, 8080])
      result = template.result(binding)
      expect(result).to eq("SERVICE: 80,443,8080")
    end

    it "maintains OpenStruct API compatibility" do
      # Test that PropertyStruct can be used as a drop-in replacement for OpenStruct
      ps = PropertyStruct.new(field1: "value1", field2: "value2")
      expect(ps).to respond_to(:field1)
      expect(ps).to respond_to(:field2)
      expect(ps.field1).to eq("value1")
      expect(ps.field2).to eq("value2")
    end
  end
end
