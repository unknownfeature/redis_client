#!/bin/bash
env_vars_options="redis_url=<host>:<port>,web_template=./src/console/vendor/redis_client/pkg/templates/console.html"

curr_dir=$(pwd)
project_name=redis_client
pkg_dir=$curr_dir/pkg
dist_dir=$curr_dir/dist
config_file=$curr_dir/configs/config.json
deploy=$1
[[ -d $dist_dir ]] && rm -rf $dist_dir || mkdir -p $dist_dir
modules=($(ls $curr_dir/pkg))
function_modules=($(ls $curr_dir/pkg/function))
for mod in "${function_modules[@]}"; do
  lower_func=$(echo $mod | tr "[:upper:]" "[:lower:]")
  . $curr_dir/pkg/function/${mod}/.function

  mkdir -p $dist_dir/$lower_func/vendor/$project_name/pkg

  cp $pkg_dir/function/$mod/* $dist_dir/$lower_func

  cd $dist_dir/$lower_func
  wire
  mv -f wire_gen.go function.go
  go mod init
  go mod tidy
  go mod vendor
  rm go.mod
  rm go.sum
  echo $modules
  for m in "${modules[@]}"; do
    [[ $m != "function" ]] && mkdir -p $dist_dir/$lower_func/vendor/$project_name/pkg/$m && cp -r $pkg_dir/$m/* $dist_dir/$lower_func/vendor/$project_name/pkg/$m
  done

  [[ $deploy ]] && gcloud functions deploy $lower_func --entry-point $function_name --runtime go113 --trigger-${trigger} --allow-unauthenticated --timeout=60 --memory=${memory}MB --vpc-connector projects/<project>/locations/us-central1/connectors/<connector> --set-env-vars $env_vars_options &
  cd $curr_dir
done
