# @AI-Generated
# Modified with AI assistance
# Description:
# 2026-05-22: Add full BOSH link support (link, if_link, EvaluationLink, EvaluationLinkInstance, ManualLinkDnsEncoder, UnknownLink) matching bosh-common interface - Cursor: Claude Sonnet 4.6

# Based on common/properties/template_evaluation_context.rb
require "rubygems"
require "json"
require "erb"
require "yaml"

# Simple struct-like class to replace OpenStruct dependency
# OpenStruct is being removed from Ruby standard library in Ruby 3.5+
class PropertyStruct
  def initialize(hash = {})
    @table = {}
    hash.each do |key, value|
      @table[key.to_sym] = wrap_value(value)
    end
  end

  def method_missing(method_name, *args)
    if method_name.to_s.end_with?("=")
      @table[method_name.to_s.chomp("=").to_sym] = wrap_value(args.first)
    else
      @table[method_name.to_sym]
    end
  end

  def respond_to_missing?(method_name, _include_private = false)
    @table.key?(method_name.to_sym) || method_name.to_s.end_with?("=")
  end

  private

  def wrap_value(value)
    case value
    when Hash
      PropertyStruct.new(value)
    when Array
      value.map { |item| wrap_value(item) }
    else
      value
    end
  end
end

class Hash
  def recursive_merge!(other)
    merge!(other) do |_, old_value, new_value|
      if old_value.class == Hash && new_value.class == Hash # rubocop:disable Style/ClassEqualityComparison
        old_value.recursive_merge!(new_value)
      else
        new_value
      end
    end
    self
  end
end

# Shared dotted-path property lookup used by TemplateEvaluationContext,
# EvaluationLink, and EvaluationLinkInstance.
module PropertyLookup
  def lookup_property(collection, name)
    keys = name.split(".")
    ref = collection
    keys.each do |key|
      ref = ref[key]
      return nil if ref.nil?
    end
    ref
  end
end

# Raised when a required link is not available.
# Matches bosh-common's UnknownLink error.
class UnknownLink < StandardError
  def initialize(name)
    super("Can't find link '#{name}'")
  end
end

# Raised when a required property is not available.
# Matches bosh-common's UnknownProperty error.
class UnknownProperty < StandardError
  attr_reader :name

  def initialize(names)
    @names = names
    super("Can't find property '#{names.join("', or '")}'")
  end
end

# Returned by if_p / if_link when the value/link was NOT found.
# Provides .else { } and .else_if_p / .else_if_link chaining.
# Matches bosh-common's ActiveElseBlock.
class ActiveElseBlock
  def initialize(context)
    @context = context
  end

  def else
    yield
  end

  def else_if_p(*names, &block) # rubocop:disable Style/ArgumentsForwarding
    @context.if_p(*names, &block) # rubocop:disable Style/ArgumentsForwarding
  end

  def else_if_link(name, &block) # rubocop:disable Style/ArgumentsForwarding
    @context.if_link(name, &block) # rubocop:disable Style/ArgumentsForwarding
  end
end

# Returned by if_p / if_link when the value/link WAS found.
# .else { } is a no-op; all else_if_* return InactiveElseBlock.
# Matches bosh-common's InactiveElseBlock.
class InactiveElseBlock
  def else
  end

  def else_if_p(*_names)
    InactiveElseBlock.new
  end

  def else_if_link(_name)
    InactiveElseBlock.new
  end
end

# Handles link.address() calls for manual links.
# When a manifest-level 'consumes' entry has an 'address' field, the resolver
# sets LinkSpec.Address so the ERB renderer knows to use this encoder.
# Matches bosh-common's ManualLinkDnsEncoder interface.
class ManualLinkDnsEncoder
  def initialize(address)
    @address = address
  end

  def encode_query(_criteria, _use_short_dns_addresses, _use_link_dns_names)
    @address
  end
end

# Represents a single instance within a resolved BOSH link.
# Wraps the per-instance hash from the LinkSpec's 'instances' array.
# Matches bosh-common's EvaluationLinkInstance interface.
class EvaluationLinkInstance
  include PropertyLookup

  attr_reader :name, :index, :id, :az, :address, :bootstrap, :properties

  def initialize(instance_spec)
    @name       = instance_spec["name"]
    @index      = instance_spec["index"]
    @id         = instance_spec["id"]
    @az         = instance_spec["az"]
    @address    = instance_spec["address"]
    @bootstrap  = instance_spec["bootstrap"]
    @properties = instance_spec["properties"] || {}
  end

  def p(*args)
    names = Array(args[0])
    names.each do |name|
      result = lookup_property(@properties, name)
      return result unless result.nil?
    end
    return args[1] if args.length == 2

    raise UnknownProperty.new(names)
  end

  def if_p(*names)
    values = names.map do |name|
      value = lookup_property(@properties, name)
      return ActiveElseBlock.new(self) if value.nil?

      value
    end
    yield(*values)
    InactiveElseBlock.new
  end

  # EvaluationLinkInstance does not have its own links; return InactiveElseBlock
  # so that else_if_link chaining from if_p blocks does not raise.
  def if_link(_name)
    InactiveElseBlock.new
  end
end

# Represents a resolved BOSH link exposed to ERB templates via link('name').
# Matches bosh-common's EvaluationLink interface exactly.
class EvaluationLink
  include PropertyLookup

  attr_reader :instances, :properties

  def initialize(link_spec, dns_encoder)
    raw_instances = link_spec["instances"] || []
    @instances   = raw_instances.map { |i| EvaluationLinkInstance.new(i) }
    @properties  = link_spec["properties"] || {}
    @dns_encoder = dns_encoder

    # Used by address() when use_link_dns_names is true.
    @group_name            = link_spec["group_name"]
    @use_link_dns_names    = link_spec["use_link_dns_names"] || false
    @use_short_dns_addresses = link_spec["use_short_dns_addresses"] || false
  end

  def p(*args)
    names = Array(args[0])
    names.each do |name|
      result = lookup_property(@properties, name)
      return result unless result.nil?
    end
    return args[1] if args.length == 2

    raise UnknownProperty.new(names)
  end

  def if_p(*names)
    values = names.map do |name|
      value = lookup_property(@properties, name)
      return ActiveElseBlock.new(self) if value.nil?

      value
    end
    yield(*values)
    InactiveElseBlock.new
  end

  # EvaluationLink does not contain other links.
  def if_link(_name)
    InactiveElseBlock.new
  end

  # Calls dns_encoder to resolve a group address (e.g. for VIP or LB).
  # For manual links the ManualLinkDnsEncoder returns the configured address.
  # For regular links in create-env (no BOSH DNS), this raises NotImplementedError
  # because the dns_encoder is nil.  Templates that only call link.instances.map(&:address)
  # never reach this code path.
  def address(criteria = {})
    raise NotImplementedError, "link.address() requires BOSH DNS which is not available in create-env; use link.instances.map(&:address) instead" if @dns_encoder.nil?

    use_short = criteria["use_short_dns_addresses"] || criteria[:use_short_dns_addresses] || @use_short_dns_addresses
    use_link_dns = criteria["use_link_dns_names"] || criteria[:use_link_dns_names] || @use_link_dns_names
    @dns_encoder.encode_query(criteria, use_short, use_link_dns)
  end
end

class TemplateEvaluationContext
  include PropertyLookup

  attr_reader :name, :index, :properties, :raw_properties, :spec

  def initialize(spec)
    @name  = spec["job"]["name"] if spec["job"].is_a?(Hash)
    @index = spec["index"]

    properties1 = if !spec["job_properties"].nil?
      spec["job_properties"]
    else
      spec["global_properties"].recursive_merge!(spec["cluster_properties"])
    end

    properties = {}
    spec["default_properties"].each do |name, value|
      copy_property(properties, properties1, name, value)
    end

    @properties     = openstruct(properties)
    @raw_properties = properties
    @spec           = openstruct(spec)

    # Initialise link support.  The full links map is keyed by job template name;
    # we select the slice for the current template so lookup is by link name only.
    all_links        = spec["links"] || {}
    job_template_name = spec["job_template_name"]
    @links = job_template_name ? (all_links[job_template_name] || {}) : {}
  end

  def get_binding
    binding
  end

  def p(*args)
    names = Array(args[0])

    names.each do |name|
      result = lookup_property(@raw_properties, name)
      return result unless result.nil?
    end

    return args[1] if args.length == 2

    raise UnknownProperty.new(names)
  end

  def if_p(*names)
    values = names.map do |name|
      value = lookup_property(@raw_properties, name)
      return ActiveElseBlock.new(self) if value.nil?

      value
    end

    yield(*values)
    InactiveElseBlock.new
  end

  # Raises UnknownLink when the link was not resolved (required link absent).
  # Matches bosh-common's EvaluationContext#link exactly.
  def link(name)
    link_spec = @links[name]
    raise UnknownLink.new(name) if link_spec.nil?
    raise UnknownLink.new(name) unless link_spec.key?("instances")

    create_evaluation_link(link_spec)
  end

  # Yields the EvaluationLink if resolved; otherwise returns ActiveElseBlock.
  # Matches bosh-common's EvaluationContext#if_link exactly.
  def if_link(name)
    link_spec = @links[name]
    if link_spec.nil? || !link_spec.key?("instances")
      return ActiveElseBlock.new(self)
    end

    yield create_evaluation_link(link_spec)
    InactiveElseBlock.new
  end

  private

  # Builds an EvaluationLink from the resolved LinkSpec hash.
  # Selects ManualLinkDnsEncoder when the spec has a top-level "address" (manual link).
  # Matches bosh-common's EvaluationContext#create_evaluation_link.
  def create_evaluation_link(link_spec)
    dns_encoder = if link_spec["address"] && !link_spec["address"].empty?
      ManualLinkDnsEncoder.new(link_spec["address"])
    else
      nil # create-env has no BOSH DNS; link.address() will raise NotImplementedError
    end
    EvaluationLink.new(link_spec, dns_encoder)
  end

  def copy_property(dst, src, name, default = nil)
    keys = name.split(".")
    src_ref = src
    dst_ref = dst

    keys.each do |key|
      src_ref = src_ref[key]
      break if src_ref.nil? # no property with this name is src
    end

    keys[0..-2].each do |key|
      dst_ref[key] ||= {}
      dst_ref = dst_ref[key]
    end

    dst_ref[keys[-1]] ||= {}
    dst_ref[keys[-1]] = src_ref.nil? ? default : src_ref
  end

  def openstruct(object)
    case object
    when Hash
      mapped = object.each_with_object({}) do |(k, v), h|
        h[k] = openstruct(v)
      end
      PropertyStruct.new(mapped)
    when Array
      object.map { |item| openstruct(item) }
    else
      object
    end
  end
end

# TODO: do not use JSON in releases
class << JSON
  alias_method :dump_array_or_hash, :dump

  def dump(*args)
    arg = args[0]
    if arg.is_a?(String) || arg.is_a?(Numeric)
      arg.inspect
    else
      dump_array_or_hash(*args)
    end
  end
end

class ERBRenderer
  def initialize(json_context_path)
    @json_context_path = json_context_path
  end

  def render(src_path, dst_path)
    erb = ERB.new(File.read(src_path), trim_mode: "-")
    erb.filename = src_path

    # NOTE: JSON.load_file was added in v2.3.1: https://github.com/ruby/json/blob/v2.3.1/lib/json/common.rb#L286
    context_hash = JSON.parse(File.read(@json_context_path))
    template_evaluation_context = TemplateEvaluationContext.new(context_hash)

    File.write(dst_path, erb.result(template_evaluation_context.get_binding))
  rescue Exception => e # rubocop:disable Lint/RescueException
    name = "#{template_evaluation_context&.name}/#{template_evaluation_context&.index}"

    line_i = e.backtrace.index { |l| l.include?(erb&.filename.to_s) }
    line_num = line_i ? e.backtrace[line_i].split(":")[1] : "unknown"
    location = "(line #{line_num}: #{e.inspect})"

    raise("Error filling in template '#{src_path}' for #{name} #{location}")
  end
end

if $0 == __FILE__
  json_context_path, erb_template_path, rendered_template_path = *ARGV

  renderer = ERBRenderer.new(json_context_path)
  renderer.render(erb_template_path, rendered_template_path)
end
