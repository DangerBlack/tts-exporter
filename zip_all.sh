cd output
for i in */; do zip -r "../zip/${i%/}.zip" "$i"; done
cd ..