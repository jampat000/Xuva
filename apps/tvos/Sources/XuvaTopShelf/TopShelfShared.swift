import Foundation

let topShelfAppGroup = "group.com.xuva.tvos"
let topShelfPayloadKey = "xuva.topShelf.payload"

struct TopShelfEntry: Codable {
    let id: String
    let title: String
    let detailKind: String
    let detailId: String
    let imageFilename: String?
    let progress: Double?
}

struct TopShelfPayload: Codable {
    let items: [TopShelfEntry]
    let sectionTitle: String
}

extension TopShelfEntry {
    var localImageURL: URL? {
        guard let filename = imageFilename,
              let container = FileManager.default.containerURL(
                forSecurityApplicationGroupIdentifier: topShelfAppGroup) else { return nil }
        return container.appendingPathComponent("topshelf-images/\(filename)")
    }

    var deepLinkURL: URL? {
        URL(string: "xuva://open/\(detailKind)/\(detailId)")
    }
}
