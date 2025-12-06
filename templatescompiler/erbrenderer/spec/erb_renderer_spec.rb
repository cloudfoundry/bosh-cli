require "spec_helper"
require "erb_renderer"

RSpec.describe "erb_renderer" do
  describe "ERBRenderer" do
    let(:test_tmpdir) { Dir.tmpdir }
    let(:json_context_path) { File.join(test_tmpdir, "context.json") }

    before do
      File.write(json_context_path, context_hash.to_json)
    end

    describe "#initialize" do
      let(:context_hash) { {} }

      it "does not raise an error" do
        expect { ERBRenderer.new(json_context_path) }.not_to raise_error
      end
    end

    describe "#render" do
      let(:context_hash) do
        {
          index: 867_5309,
          global_properties: {
            property1: "global_value1"
          },
          cluster_properties: {
            property2: "cluster_value1"
          },
          default_properties: {
            property1: "default_value1",
            property2: "default_value2",
            property3: "default_value3"
          }
        }
      end

      let(:erb_template_path) { File.join(test_tmpdir, "template.yml.erb") }
      let(:erb_content) do
        <<~TEST_TEMPLATE
          ---
          property1: <%= p('property1') %>
          property2: <%= p('property2') %>
          property3: <%= p('property3') %>
        TEST_TEMPLATE
      end

      let(:rendered_template_path) { File.join(test_tmpdir, "template.yml") }
      let(:expected_template_content) do
        <<~EXPECTED_TEMPLATE
          ---
          property1: global_value1
          property2: cluster_value1
          property3: default_value3
        EXPECTED_TEMPLATE
      end

      let(:erb_renderer) { ERBRenderer.new(json_context_path) }

      before do
        File.write(erb_template_path, erb_content)
      end

      it "does not raise an error" do
        expect { erb_renderer.render(erb_template_path, rendered_template_path) }.not_to raise_error
      end

      it "renders the expected content" do
        erb_renderer.render(erb_template_path, rendered_template_path)
        expect(File.read(rendered_template_path)).to eq(expected_template_content)
      end
    end
  end
end
