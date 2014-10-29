package erbrenderer

const templateEvaluationContextRb = `
# Based on common/properties/template_evaluation_context.rb
require "rubygems"
require "ostruct"
require "json"
require "erb"

class TemplateEvaluationContext
  attr_reader :name, :index
  attr_reader :properties, :raw_properties
  attr_reader :spec

  def initialize(spec)
    @name = spec["job"]["name"] if spec["job"].is_a?(Hash)
    @index = spec["index"]
    @properties = openstruct(spec["properties"] || {})
    @raw_properties = spec["properties"] || {}
    @spec = openstruct(spec)
  end

  def get_binding
    binding.taint
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

    yield *values
    InactiveElseBlock.new
  end

  private

  def openstruct(object)
    case object
      when Hash
        mapped = object.inject({}) { |h, (k,v)| h[k] = openstruct(v); h }
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

    def initialize(name)
      @name = name
      super("Can't find property '#{name}'")
    end
  end

  class ActiveElseBlock
    def initialize(template)
      @context = template
    end

    def else
      yield
    end

    def else_if_p(*names, &block)
      @context.if_p(*names, &block)
    end
  end

  class InactiveElseBlock
    def else; end

    def else_if_p(*names)
      InactiveElseBlock.new
    end
  end
end

# todo do not use JSON in releases
class << JSON
  alias dump_array_or_hash dump

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
  def initialize(context)
    @context = context
  end

  def render(src_path, dst_path)
    erb = ERB.new(File.read(src_path))
    erb.filename = src_path

    File.open(dst_path, "w") do |f|
      f.write(erb.result(@context.get_binding))
    end

  rescue Exception => e
    name = "#{@context.name}/#{@context.index}"

    line_i = e.backtrace.index { |l| l.include?(erb.filename) }
    line_num = line_i ? e.backtrace[line_i].split(':')[1] : "unknown"
    location = "(line #{line_num}: #{e.inspect})"

    raise("Error filling in template '#{src_path}' for #{name} #{location}")
  end
end

if $0 == __FILE__
  context_path, src_path, dst_path = *ARGV

  context_hash = JSON.load(File.read(context_path))
  context = TemplateEvaluationContext.new(context_hash)

  renderer = ERBRenderer.new(context)
  renderer.render(src_path, dst_path)
end
`
