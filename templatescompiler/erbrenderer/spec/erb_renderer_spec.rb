require "spec_helper"
require "erb_renderer"

RSpec.describe "erb_renderer" do
  describe "ERBRenderer" do
    let(:json_context_path) { File.join(Dir.tmpdir, "context.json") }

    describe "#initialize" do
      before do
        File.write(json_context_path, {}.to_json)
      end

      it "does not raise an error" do
        expect { ERBRenderer.new(json_context_path) }.not_to raise_error
      end
    end
  end
end
