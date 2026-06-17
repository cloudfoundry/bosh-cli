# @AI-Generated
# Modified with AI assistance
# Description:
# 2026-05-22: Add RSpec tests for BOSH link support in erb_renderer.rb - Cursor: Claude Sonnet 4.6

require "spec_helper"
require "erb_renderer"

RSpec.describe "BOSH link support in erb_renderer" do
  let(:test_tmpdir)      { Dir.tmpdir }
  let(:json_context_path) { File.join(test_tmpdir, "context.json") }
  let(:erb_template_path) { File.join(test_tmpdir, "template.erb") }
  let(:rendered_path)     { File.join(test_tmpdir, "output.txt") }
  let(:erb_renderer)      { ERBRenderer.new(json_context_path) }

  # Minimal base context that satisfies TemplateEvaluationContext#initialize.
  let(:base_context) do
    {
      "index"              => 0,
      "job"                => {"name" => "my-job"},
      "global_properties"  => {},
      "cluster_properties" => {},
      "default_properties" => {},
      "job_template_name"  => "my-job",
      "links"              => {}
    }
  end

  before do
    File.write(json_context_path, context.to_json)
    File.write(erb_template_path, erb_content)
  end

  after do
    [json_context_path, erb_template_path, rendered_path].each do |f|
      File.unlink(f) if File.exist?(f)
    end
  end

  def render
    erb_renderer.render(erb_template_path, rendered_path)
    File.read(rendered_path)
  end

  # ---------------------------------------------------------------------------
  # link() — resolved link
  # ---------------------------------------------------------------------------
  describe "link()" do
    let(:context) do
      base_context.merge(
        "links" => {
          "my-job" => {
            "mysql" => {
              "deployment_name"        => "bosh",
              "domain"                 => "bosh",
              "instance_group"         => "bosh",
              "default_network"        => "default",
              "group_name"             => "mysql.bosh.bosh.bosh",
              "instances"              => [
                {"name" => "bosh", "id" => "abc", "index" => 0, "bootstrap" => true,
                 "az" => "z1", "address" => "10.0.0.1"},
                {"name" => "bosh", "id" => "def", "index" => 1, "bootstrap" => false,
                 "az" => "z1", "address" => "10.0.0.2"}
              ],
              "properties"             => {"port" => 13306},
              "use_link_dns_names"     => false,
              "use_short_dns_addresses" => false
            }
          }
        }
      )
    end

    context "accessing instances" do
      let(:erb_content) do
        '<%= link("mysql").instances.map(&:address).join(",") %>'
      end

      it "returns all instance addresses" do
        expect(render).to eq("10.0.0.1,10.0.0.2")
      end
    end

    context "accessing instance fields" do
      let(:erb_content) do
        '<%= link("mysql").instances.first.index %>'
      end

      it "exposes instance index" do
        expect(render).to eq("0")
      end

      it "exposes instance bootstrap flag" do
        File.write(erb_template_path, '<%= link("mysql").instances.first.bootstrap %>')
        expect(render).to eq("true")
      end

      it "exposes instance id" do
        File.write(erb_template_path, '<%= link("mysql").instances.first.id %>')
        expect(render).to eq("abc")
      end

      it "exposes instance az" do
        File.write(erb_template_path, '<%= link("mysql").instances.first.az %>')
        expect(render).to eq("z1")
      end
    end

    context "accessing link properties via p()" do
      let(:erb_content) { '<%= link("mysql").p("port") %>' }

      it "returns the link property value" do
        expect(render).to eq("13306")
      end
    end

    context "accessing link properties via p() with default" do
      let(:erb_content) { '<%= link("mysql").p("nonexistent", 9999) %>' }

      it "returns the default when property is missing" do
        expect(render).to eq("9999")
      end
    end

    context "using if_p() on a link" do
      let(:erb_content) do
        <<~ERB.chomp
          <% link("mysql").if_p("port") do |port| %>port=<%= port %><% end %>
        ERB
      end

      it "yields the property value" do
        expect(render).to eq("port=13306")
      end
    end

    context "using if_p() with absent property on a link" do
      let(:erb_content) do
        <<~ERB.chomp
          <% link("mysql").if_p("nonexistent") do |v| %>YES<% end.else do %>NO<% end %>
        ERB
      end

      it "executes the else block" do
        expect(render).to eq("NO")
      end
    end

    context "when the link does not exist" do
      let(:erb_content) { '<%= link("no-such-link").instances.first.address %>' }

      it "raises an UnknownLink error" do
        expect { render }.to raise_error(/Can't find link 'no-such-link'/)
      end
    end
  end

  # ---------------------------------------------------------------------------
  # if_link() — present link
  # ---------------------------------------------------------------------------
  describe "if_link() with a resolved link" do
    let(:context) do
      base_context.merge(
        "links" => {
          "my-job" => {
            "mysql" => {
              "deployment_name"        => "bosh",
              "domain"                 => "bosh",
              "instance_group"         => "bosh",
              "default_network"        => "default",
              "group_name"             => "mysql.bosh.bosh.bosh",
              "instances"              => [
                {"name" => "bosh", "id" => "abc", "index" => 0, "bootstrap" => true,
                 "az" => "z1", "address" => "10.0.0.1"}
              ],
              "properties"             => {"port" => 3306},
              "use_link_dns_names"     => false,
              "use_short_dns_addresses" => false
            }
          }
        }
      )
    end

    context "yields the EvaluationLink" do
      let(:erb_content) do
        <<~ERB.chomp
          <% if_link("mysql") do |mysql| %>addr=<%= mysql.instances.first.address %><% end %>
        ERB
      end

      it "renders the block content" do
        expect(render).to eq("addr=10.0.0.1")
      end
    end

    context ".else block is not executed when link is present" do
      let(:erb_content) do
        <<~ERB.chomp
          <% if_link("mysql") do |_| %>found<% end.else do %>not-found<% end %>
        ERB
      end

      it "does not render the else block" do
        expect(render).to eq("found")
      end
    end
  end

  # ---------------------------------------------------------------------------
  # if_link() — absent link → else branch
  # ---------------------------------------------------------------------------
  describe "if_link() with an absent link" do
    let(:context) { base_context }

    context ".else block executes" do
      let(:erb_content) do
        <<~ERB.chomp
          <% if_link("no-link") do |_| %>found<% end.else do %>not-found<% end %>
        ERB
      end

      it "renders the else block" do
        expect(render).to eq("not-found")
      end
    end

    context ".else_if_link with another absent link" do
      let(:erb_content) do
        <<~ERB.chomp
          <% if_link("no-link") do |_| %>A<% end.else_if_link("also-no-link") do |_| %>B<% end.else do %>C<% end %>
        ERB
      end

      it "falls through to the final else" do
        expect(render).to eq("C")
      end
    end

    context ".else_if_p chains from if_link absent block" do
      let(:context) do
        base_context.merge(
          "global_properties"  => {"fallback" => "fallback-value"},
          "default_properties" => {"fallback" => nil}
        )
      end

      let(:erb_content) do
        <<~ERB.chomp
          <% if_link("no-link") do |_| %>link<% end.else_if_p("fallback") do |v| %><%= v %><% end %>
        ERB
      end

      it "executes else_if_p when link is absent" do
        expect(render).to eq("fallback-value")
      end
    end
  end

  # ---------------------------------------------------------------------------
  # else_if_link chains from if_link present block (InactiveElseBlock)
  # ---------------------------------------------------------------------------
  describe "else_if_link on InactiveElseBlock" do
    let(:context) do
      base_context.merge(
        "links" => {
          "my-job" => {
            "mysql" => {
              "deployment_name"        => "bosh",
              "domain"                 => "bosh",
              "instance_group"         => "bosh",
              "default_network"        => "default",
              "group_name"             => "",
              "instances"              => [
                {"name" => "bosh", "id" => "a", "index" => 0, "bootstrap" => true,
                 "az" => "", "address" => "10.0.0.1"}
              ],
              "properties"             => {},
              "use_link_dns_names"     => false,
              "use_short_dns_addresses" => false
            }
          }
        }
      )
    end

    let(:erb_content) do
      <<~ERB.chomp
        <% if_link("mysql") do |_| %>primary<% end.else_if_link("other") do |_| %>secondary<% end %>
      ERB
    end

    it "does not execute else_if_link when primary link was found" do
      expect(render).to eq("primary")
    end
  end

  # ---------------------------------------------------------------------------
  # Manual link — address field triggers ManualLinkDnsEncoder
  # ---------------------------------------------------------------------------
  describe "manual link with address field" do
    let(:context) do
      base_context.merge(
        "links" => {
          "my-job" => {
            "db" => {
              "deployment_name"        => "bosh",
              "domain"                 => "bosh",
              "instance_group"         => "bosh",
              "default_network"        => "default",
              "group_name"             => "",
              "instances"              => [
                {"name" => "external", "id" => "manual-0", "index" => 0, "bootstrap" => true,
                 "az" => "", "address" => "192.168.1.100"}
              ],
              "properties"             => {"port" => 3306},
              "address"                => "192.168.1.100",
              "use_link_dns_names"     => false,
              "use_short_dns_addresses" => false
            }
          }
        }
      )
    end

    context "link.address() returns the fixed address" do
      let(:erb_content) { '<%= link("db").address %>' }

      it "returns the manually configured address" do
        expect(render).to eq("192.168.1.100")
      end
    end

    context "link.instances[].address still works" do
      let(:erb_content) { '<%= link("db").instances.first.address %>' }

      it "returns the per-instance address" do
        expect(render).to eq("192.168.1.100")
      end
    end
  end

  # ---------------------------------------------------------------------------
  # link.address() raises NotImplementedError without DNS encoder
  # ---------------------------------------------------------------------------
  describe "link.address() without a manual link (no address field)" do
    let(:context) do
      base_context.merge(
        "links" => {
          "my-job" => {
            "mysql" => {
              "deployment_name"        => "bosh",
              "domain"                 => "bosh",
              "instance_group"         => "bosh",
              "default_network"        => "default",
              "group_name"             => "",
              "instances"              => [
                {"name" => "bosh", "id" => "a", "index" => 0, "bootstrap" => true,
                 "az" => "", "address" => "10.0.0.1"}
              ],
              "properties"             => {},
              "use_link_dns_names"     => false,
              "use_short_dns_addresses" => false
            }
          }
        }
      )
    end

    let(:erb_content) { '<%= link("mysql").address %>' }

    it "raises an error mentioning create-env" do
      # ERBRenderer#render rescues all exceptions and re-raises as RuntimeError,
      # so the underlying NotImplementedError message is wrapped.
      expect { render }.to raise_error(RuntimeError, /create-env/)
    end
  end

  # ---------------------------------------------------------------------------
  # job_template_name scoping — each job sees only its own links
  # ---------------------------------------------------------------------------
  describe "link scoping by job_template_name" do
    let(:context) do
      base_context.merge(
        "job_template_name" => "consumer-job",
        "links" => {
          "consumer-job" => {
            "mysql" => {
              "deployment_name"        => "bosh",
              "domain"                 => "bosh",
              "instance_group"         => "bosh",
              "default_network"        => "default",
              "group_name"             => "",
              "instances"              => [
                {"name" => "bosh", "id" => "a", "index" => 0, "bootstrap" => true,
                 "az" => "", "address" => "10.0.0.1"}
              ],
              "properties"             => {},
              "use_link_dns_names"     => false,
              "use_short_dns_addresses" => false
            }
          },
          "other-job" => {
            "mysql" => {
              "deployment_name"        => "bosh",
              "domain"                 => "bosh",
              "instance_group"         => "bosh",
              "default_network"        => "default",
              "group_name"             => "",
              "instances"              => [
                {"name" => "bosh", "id" => "z", "index" => 0, "bootstrap" => true,
                 "az" => "", "address" => "10.0.0.99"}
              ],
              "properties"             => {},
              "use_link_dns_names"     => false,
              "use_short_dns_addresses" => false
            }
          }
        }
      )
    end

    let(:erb_content) { '<%= link("mysql").instances.first.address %>' }

    it "uses only the current job's link, not another job's link with the same name" do
      expect(render).to eq("10.0.0.1")
    end
  end

  # ---------------------------------------------------------------------------
  # EvaluationLinkInstance p() and if_p()
  # ---------------------------------------------------------------------------
  describe "EvaluationLinkInstance p() and if_p()" do
    let(:context) do
      base_context.merge(
        "links" => {
          "my-job" => {
            "db" => {
              "deployment_name"        => "bosh",
              "domain"                 => "bosh",
              "instance_group"         => "bosh",
              "default_network"        => "default",
              "group_name"             => "",
              "instances"              => [
                {"name" => "bosh", "id" => "a", "index" => 0, "bootstrap" => true,
                 "az" => "", "address" => "10.0.0.1",
                 "properties" => {"role" => "primary"}}
              ],
              "properties"             => {},
              "use_link_dns_names"     => false,
              "use_short_dns_addresses" => false
            }
          }
        }
      )
    end

    let(:erb_content) do
      '<%= link("db").instances.first.p("role") %>'
    end

    it "reads per-instance properties" do
      expect(render).to eq("primary")
    end

    context "if_p on instance" do
      let(:erb_content) do
        <<~ERB.chomp
          <% link("db").instances.first.if_p("role") do |r| %>role=<%= r %><% end %>
        ERB
      end

      it "yields the per-instance property" do
        expect(render).to eq("role=primary")
      end
    end
  end
end
