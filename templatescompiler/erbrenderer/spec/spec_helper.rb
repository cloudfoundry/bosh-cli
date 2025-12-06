require "rspec"
require "json"
require "tmpdir"

ERB_RENDERER_ROOT = File.expand_path("..", File.dirname(__FILE__))

$LOAD_PATH.unshift(ERB_RENDERER_ROOT)
