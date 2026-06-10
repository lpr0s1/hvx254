import formidable from "formidable";
import fs from "fs";

export const config = {
  api: { bodyParser: false }
};

let lastMessage = "Aucun message pour le moment";
let lastFile = null;

export default function handler(req, res) {
  if (req.method === "GET") {
    return res.status(200).json({
      status: "Disponible",
      message: lastMessage,
      file: lastFile
    });
  }

  if (req.method === "POST") {
    const form = formidable({ multiples: false });

    form.parse(req, (err, fields, files) => {
      if (err) {
        return res.status(500).json({ error: "Erreur lors du parsing du fichier" });
      }

      // Message texte
      if (fields.message) {
        lastMessage = fields.message;
      }

      // Fichier
      if (files.file) {
        const file = files.file;
        const buffer = fs.readFileSync(file.filepath);

        lastFile = {
          filename: file.originalFilename,
          mimetype: file.mimetype,
          size: file.size,
          base64: buffer.toString("base64")
        };
      }

      return res.status(200).json({
        status: "OK",
        receivedMessage: lastMessage,
        receivedFile: lastFile ? lastFile.filename : null
      });
    });
  }

  return res.status(405).json({ error: "405" });
}
