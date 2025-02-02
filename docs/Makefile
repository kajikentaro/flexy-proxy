gen:
	rm -rf output
	mkdir output
	docker build . -t gen-schema-doc
	cp ../models/config-spec.json ./
	docker run -it --rm -w /app/output -v .:/app gen-schema-doc generate-schema-doc ../ index.html