# Based on common/properties/template_evaluation_context.rb
require "rubygems"
require "ostruct"
require "json"
require "erb"
require "yaml"

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

class TemplateEvaluationContext
  attr_reader :name, :index
  attr_reader :properties, :raw_properties
  attr_reader :spec

  def initialize(spec)
    @name = spec["job"]["name"] if spec["job"].is_a?(Hash)
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

    @properties = openstruct(properties)
    @raw_properties = properties
    @spec = openstruct(spec)
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

  def if_link(name)
    false
  end

  private

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
      mapped = object.each_with_object({}) { |(k, v), h|
        h[k] = openstruct(v)
      }
      OpenStruct.new(mapped)
    when Array
      object.map { |item| openstruct(item) }
    else
      object
    end
  end

  def lookup_property(collection, name)
    keys = name.split(".")
    ref = collection

    keys.each do |key|
      ref = ref[key]
      return nil if ref.nil?
    end

    ref
  end

  class UnknownProperty < StandardError
    attr_reader :name

    def initialize(names)
      @names = names
      super("Can't find property '#{names.join("', or '")}'")
    end
  end

  class ActiveElseBlock
    def initialize(template)
      @context = template
    end

    def else
      yield
    end

    def else_if_p(*names, &block) # rubocop:disable Style/ArgumentsForwarding
      @context.if_p(*names, &block) # rubocop:disable Style/ArgumentsForwarding
    end
  end

  class InactiveElseBlock
    def else
    end

    def else_if_p(*names)
      InactiveElseBlock.new
    end
  end
end

# todo do not use JSON in releases
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

    # Note: JSON.load_file was added in v2.3.1: https://github.com/ruby/json/blob/v2.3.1/lib/json/common.rb#L286
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
