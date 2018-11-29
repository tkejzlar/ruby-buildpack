class WelcomesController < ApplicationController
  def index
    @asset_contents = Dir.entries("#{Rails.root}/public/assets/")
    @js_asset_count = @asset_contents.select { |file| file.include?('javaScriptAsset') }.size
    @welcome = 'Hello World'
  end
end
