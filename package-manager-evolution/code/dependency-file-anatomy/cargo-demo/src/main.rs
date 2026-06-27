use itertools::Itertools;

fn main() {
    let joined = ["lock", "diff", "matters"].iter().join("-");
    println!("{}", joined);
}
