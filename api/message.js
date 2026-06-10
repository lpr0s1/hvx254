let lastMessage = "Aucun message pour le moment";
let lastFile = null;

export const config = {
  api: { bodyParser: true }
};

export default async function handler(req, res) {
  // Télécharger le fichier
  if (req.method === "GET" && req.url.endsWith("/file")) {
    if (!lastFile) return res.status(404).send("Aucun fichier stocké");

    const buffer = Buffer.from(lastFile.base64, "base64");
    res.setHeader("Content-Type", "application/octet-stream");
    res.setHeader("Content-Disposition", `attachment; filename="${lastFile.filename}"`);
    return res.send(buffer);
  }

  // GET normal
  if (req.method === "GET") {
    return res.status(200).json({
      status: "Disponible",
      message: lastMessage,
      file: lastFile
    });
  }

  // POST JSON
  if (req.method === "POST") {
    const { message, filename, base64 } = req.body;

    if (message) lastMessage = message;

    if (filename && base64) {
      lastFile = { filename, base64 };
    }

    return res.status(200).json({
      status: "OK",
      receivedMessage: lastMessage,
      receivedFile: lastFile ? lastFile.filename : null
    });
  }

  return res.status(405).json({ error: "405" });
}
