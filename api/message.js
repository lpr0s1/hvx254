let lastMessage = "Aucun message pour le moment";
let lastFile = null;

export const config = {
  api: {
    bodyParser: false
  }
};

export default async function handler(req, res) {
  if (req.method === "GET") {
    return res.status(200).json({
      status: "Disponible",
      message: lastMessage,
      file: lastFile
    });
  }

  if (req.method === "POST") {
    const chunks = [];
    for await (const chunk of req) {
      chunks.push(chunk);
    }

    const buffer = Buffer.concat(chunks);

    // Vérifier si c'est un multipart/form-data
    const contentType = req.headers["content-type"] || "";

    if (contentType.includes("multipart/form-data")) {
      const boundary = contentType.split("boundary=")[1];
      const parts = buffer.toString().split(`--${boundary}`);

      for (const part of parts) {
        if (part.includes("Content-Disposition")) {
          // Fichier
          if (part.includes("filename=")) {
            const filename = part.match(/filename="(.+?)"/)[1];
            const fileContent = part.split("\r\n\r\n")[1].split("\r\n")[0];

            lastFile = {
              filename,
              base64: Buffer.from(fileContent, "binary").toString("base64")
            };
          }

          // Message texte
          if (part.includes('name="message"')) {
            const msg = part.split("\r\n\r\n")[1].split("\r\n")[0];
            lastMessage = msg;
          }
        }
      }

      return res.status(200).json({
        status: "OK",
        receivedMessage: lastMessage,
        receivedFile: lastFile ? lastFile.filename : null
      });
    }

    return res.status(400).json({ error: "Format non supporté" });
  }

  return res.status(405).json({ error: "405" });
}
